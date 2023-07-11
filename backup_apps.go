package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/apigee/v1"
	"google.golang.org/api/option"
)

type Attribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AppBackup struct {
	AppID       string      `json:"appId"`
	Attributes  []Attribute `json:"attributes"`
	CreatedAt   int64       `json:"createdAt"`
	Credentials []struct {
		APIProducts []struct {
			APIProduct string `json:"apiproduct"`
			Status     string `json:"status"`
		} `json:"apiProducts"`
		ConsumerKey    string `json:"consumerKey"`
		ConsumerSecret string `json:"consumerSecret"`
		ExpiresAt      string `json:"expiresAt"`
		IssuedAt       string `json:"issuedAt"`
		Status         string `json:"status"`
	} `json:"credentials"`
	DeveloperID    string `json:"developerId"`
	LastModifiedAt int64  `json:"lastModifiedAt"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	AppFamily      string `json:"appFamily"`
}

func help() {
	fmt.Println("Usage: go run main.go <serviceAccountFile> <organization> <backupDir>")
	fmt.Println("\nDescription: Este programa faz backup de todos os Apps do Apigee.")
	fmt.Println("\n- Options: <serviceAccountFile> - Arquivo json do service account")
	fmt.Println("- Options: <organization> - Organizacao Apigee")
	fmt.Println("- Options: <backupDir> - Diretorio que deseja criar. OBS: O script cria no final do diretorio _timestamp")
	fmt.Println("\nEx: go run main.go service-account.json my-org backups")
}

func main() {

	if len(os.Args) < 4 {
		help()
		return

	}

	serviceAccountFile := os.Args[1]
	org := os.Args[2]
	backupDir := os.Args[3]

	timestamp := time.Now().Unix()

	dirBackup := backupDir + "_" + strconv.FormatInt(timestamp, 10)

	err := os.Mkdir(dirBackup, 0755)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Diretorio '%s' criado com sucesso.", backupDir)

	//serviceAccountFile := "../../credentials/admapi.json"

	ctx := context.Background()

	serviceAccountJSON, err := os.ReadFile(serviceAccountFile)
	if err != nil {
		log.Fatalf("Erro ao carregar as credenciais de Service Account %v", err)
	}

	credentials, err := google.CredentialsFromJSON(ctx, serviceAccountJSON, apigee.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Erro ao carregar as credenciais da Service Account: %v", err)
	}

	service, err := apigee.NewService(ctx, option.WithCredentials(credentials))
	if err != nil {
		log.Fatalf("Erro ao criar o cliente do Apigee: %v", err)
	}

	developers, err := service.Organizations.Developers.List("organizations/" + org).Do()
	if err != nil {
		log.Fatalf("Erro ao obter a lista de developers: %v", err)
	}

	// Percorre a lista de apps que pertencem ao developers
	for _, developer := range developers.Developer {
		apps, err := service.Organizations.Developers.Apps.List("organizations/" + org + "/developers/" + developer.Email).Do()
		if err != nil {
			log.Fatalf("Erro ao obter a lista de apps do developer %s: %v", developer.Email, err)
		}

		for _, app := range apps.App {
			appDetails, err := service.Organizations.Developers.Apps.Get("organizations/" + org + "/developers/" + developer.Email + "/apps/" + app.AppId).Do()
			if err != nil {
				log.Printf("Erro ao obter os detalhes do App: %v", err)
				continue

			}

			var attributes []Attribute
			for _, attr := range appDetails.Attributes {
				attributes = append(attributes, Attribute{
					Name:  attr.Name,
					Value: attr.Value,
				})
			}
			var credentials []struct {
				APIProducts []struct {
					APIProduct string `json:"apiproduct"`
					Status     string `json:"status"`
				} `json:"apiProducts"`
				ConsumerKey    string `json:"consumerKey"`
				ConsumerSecret string `json:"consumerSecret"`
				ExpiresAt      string `json:"expiresAt"`
				IssuedAt       string `json:"issuedAt"`
				Status         string `json:"status"`
			}

			for _, cred := range appDetails.Credentials {
				expiresAt := strconv.FormatInt(cred.ExpiresAt, 10)
				issuedAt := strconv.FormatInt(cred.IssuedAt, 10)

				var apiProducts []struct {
					APIProduct string `json:"apiproduct"`
					Status     string `json:"status"`
				}
				for _, product := range cred.ApiProducts {
					apiProducts = append(apiProducts, struct {
						APIProduct string `json:"apiproduct"`
						Status     string `json:"status"`
					}{
						APIProduct: product.Apiproduct,
						Status:     product.Status,
					})
				}

				credentials = append(credentials, struct {
					APIProducts []struct {
						APIProduct string `json:"apiproduct"`
						Status     string `json:"status"`
					} `json:"apiProducts"`
					ConsumerKey    string `json:"consumerKey"`
					ConsumerSecret string `json:"consumerSecret"`
					ExpiresAt      string `json:"expiresAt"`
					IssuedAt       string `json:"issuedAt"`
					Status         string `json:"status"`
				}{
					APIProducts:    apiProducts,
					ConsumerKey:    cred.ConsumerKey,
					ConsumerSecret: cred.ConsumerSecret,
					ExpiresAt:      expiresAt,
					IssuedAt:       issuedAt,
					Status:         cred.Status,
				})
			}

			appBackup := AppBackup{
				AppID:          app.AppId,
				Attributes:     attributes,
				CreatedAt:      app.CreatedAt,
				Credentials:    credentials,
				DeveloperID:    app.DeveloperId,
				LastModifiedAt: app.LastModifiedAt,
				Name:           app.Name,
				Status:         app.Status,
				AppFamily:      app.AppFamily,
			}

			jsonData, err := json.MarshalIndent(appBackup, "", "  ")
			if err != nil {
				log.Printf("Erro ao converter o App para JSON: %v", err)
				continue
			}

			filename := fmt.Sprintf(dirBackup+"/%s.json", app.AppId)
			err = saveToFile(filename, jsonData)
			if err != nil {
				log.Printf("Erro ao salvar o arquivo JSON: %v", err)
			}
		}
	}
}

func saveToFile(filename string, data []byte) error {
	// Abre o arquivo no modo de escrita
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo: %v", err)
	}
	defer file.Close()

	// Escreve os dados no arquivo
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("erro ao escrever no arquivo: %v", err)
	}

	return nil
}
