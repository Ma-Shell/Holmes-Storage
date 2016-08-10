package objStorerGeneric

/*
This file contains structs to represent all default
collections and interfaces.
If you need to extend these for your database specific
implementation add a wrapper to your specific storer.../*.go file,
don't change these structs here!
*/

type ObjDBConnector struct {
	IP           string
	Port         int
	Region       string
	Key          string
	Secret       string
	Bucket       string
	ConfigBucket string
	DisableSSL   bool
}

type ObjStorer interface {
	// Initializes the connection and the objStorer object
	// ip, port, user, passwort, db name
	Initialize([]*ObjDBConnector) (ObjStorer, error)

	// Is called to setup the storage system on the very first run
	// to create initial collections (if necessary)
	Setup() error

	// Stores a new sample in the database
	// return "duplicate" error if already known
	StoreSample(*Sample) error

	// Gets a sample from the database, identified
	// by its sha2 string
	GetSample(string) (*Sample, error)

	// Stores a configuration file
	StoreConfig(*Config) error

	// Gets a configuration file
	GetConfig(string) (*Config, error)
}

// TODO: switch from json to probably raw bytes
type Sample struct {
	SHA256 string `json:"sha256"`
	Data   []byte `json:"data"` //this will result in a base64 encoded string when marshaled
}

type Config struct {
	Path         string `json:"path"`
	FileContents []byte `json:"file_contents"`
}
