# Backup dos apps do Apigee

O service account precisa ter permissoes no projeto para conseguir efetuar o backup de todos os  recursos como por exemplo key e secret

No diretorio corrente executar o comando abaixo para criar o pacote \
`go init nome_do_pacote`

Executar os comandos go get nos pacotes da google \
`go get -u golang.org/x/oauth2/google` \
`go get -u google.golang.org/api/apigee/v1` \
`go mod tidy`\


# Para usar o codigo

Usage: go run backup_apps.go <serviceAccountFile> <organization> <backupDir>

Description: Este programa faz backup de todos os Apps do Apigee.

- Options: <serviceAccountFile> - Arquivo json do service account
- Options: <organization> - Organizacao Apigee
- Options: <backupDir> - Diretorio que deseja criar. OBS: O script cria no final do diretorio  _timestamp

Ex: go run backup_apps.go service-account.json my-org backups
