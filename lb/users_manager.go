package main



type User struct {
	Name    string
	AuthenticaionType   string
	certManager *certManager
}


func (u *User) LoadUsersConfig(configPath string) string {

}

func (u *User) CreateUser(name string) {
	u.certManager.CreateClientCert(name + ".crt", name + ".key")
	// Add user to users.json
}

func (u *User) DeleteUser(name string) {

}



/*

- create a user
- delete a user
- list all users | use metadata.json for this
- watch for user expiration


example structure

/etc/myproxy/
├── ca/
│   ├── ca-cert.pem
│   └── ca-key.pem
├── clients/
│   ├── alice/
│   │   ├── cert.pem
│   │   └── key.pem
│   ├── bob/
│   │   ├── cert.pem
│   │   └── key.pem


*/