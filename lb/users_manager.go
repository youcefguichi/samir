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
}

func (u *User) DeleteUser(name string) {

}



/*

- create a user
- delete a user
- list all users
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