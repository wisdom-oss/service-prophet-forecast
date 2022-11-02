package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	e "microservice/errors"
)

var logger = log.WithFields(
	log.Fields{
		"package": "helpers",
	},
)

/*
PingHost

Use this function to ping a host on a port by connecting to it via a tcp connection.

Parameters:
	- host: The host which is the target of the connection
	- port: The port which is the target of the connection
	- timeout: The connection timeout in seconds

The function returns true if the connection was successful. Else it will return false
*/
func PingHost(host string, port int, timeout int) bool {
	// Build the connection timeout
	connectionTimeout := time.Duration(timeout) * time.Second
	// Build the tcp connection target string
	connectionTarget := fmt.Sprintf("%s:%d", host, port)
	// Open a tcp connection with the parameters
	_, connectionError := net.DialTimeout("tcp", connectionTarget, connectionTimeout)
	if connectionError != nil {
		return false
	} else {
		return true
	}

}

/*
GetLocalIP

Get the local ip address as a string.
*/
func GetLocalIP() string {
	interfaceAddresses, err := net.InterfaceAddrs()
	if err != nil {
		log.WithError(err).Error("Unable to access the network interface addresses. Please check your permissions.")
		return ""
	}
	for _, address := range interfaceAddresses {
		if address, ok := address.(*net.IPNet); ok && !address.IP.IsLoopback() {
			if address.IP.To4() != nil {
				return address.IP.String()
			}
		}
	}
	return ""
}

/*
StringArrayContains

Check if the string array contains the string s as item
*/
func StringArrayContains(a []string, s string) bool {
	for _, item := range a {
		if item == s {
			return true
		}
	}
	return false
}

/*
SendRequestError

Send a new request error using the http.ResponseWriter supplied to the function
*/
func SendRequestError(errorCode string, w http.ResponseWriter) {
	requestError := e.NewRequestError(errorCode)
	w.Header().Set("Content-Type", "text/json")
	w.WriteHeader(requestError.HttpStatus)
	encodingError := json.NewEncoder(w).Encode(requestError)
	if encodingError != nil {
		logger.WithError(encodingError).Error("Unable to encode the request error into json")
		return
	}
}

// ReadEnvironmentVariable checks if the specified environment variable is populated.
// If not an error is returned with an empty string
func ReadEnvironmentVariable(envName string) (string, error) {
	val, isSet := os.LookupEnv(envName)
	if !isSet {
		return "", errors.New("environment not populated")
	} else {
		return val, nil
	}
}
