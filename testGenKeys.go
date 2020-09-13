package main

import (
	"log"

	"go_systems/src/demo2Config"
	"go_systems/src/demo2KeyGen"
	"go_systems/src/demo2fs"
)

func main() {
	key, cert, err := demo2KeyGen.GenerateKeyPair(1024, demo2Config.PKPass)
	if err != nil {
		log.Fatalf("Error generating key pair %s\n", err)
	}
	f, err := demo2fs.CreateFile(demo2Config.KeyCertPath, "mykey.pem")
	if err != nil {
		log.Fatalf("Error creating PEM file %s\n", err)
	}
	if err := demo2fs.WriteFile(f, key); err != nil {
		log.Fatalf("Error writing file %s\n", err)
	}
	f, err = demo2fs.CreateFile(demo2Config.KeyCertPath, "mykey.pub")
	if err != nil {
		log.Fatalf("Error creating Pub file %s\n", err)
	}
	if err := demo2fs.WriteFile(f, cert); err != nil {
		log.Fatalf("Error writing file %s\n", err)
	}
}
