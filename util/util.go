package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/csv"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/gophish/gophish/models"
	"github.com/jordan-wright/email"
)

// Logger is used to send logging messages to stdout.
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

// ParseMail takes in an HTTP Request and returns an Email object
// TODO: This function will likely be changed to take in a []byte
func ParseMail(r *http.Request) (email.Email, error) {
	e := email.Email{}
	m, err := mail.ReadMessage(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	body, err := ioutil.ReadAll(m.Body)
	e.HTML = body
	return e, err
}

// ParseCSV contains the logic to parse the user provided csv file containing Target entries
func ParseCSV(r *http.Request) ([]models.Target, error) {
	mr, err := r.MultipartReader()
	ts := []models.Target{}
	if err != nil {
		return ts, err
	}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		// Skip the "submit" part
		if part.FileName() == "" {
			continue
		}
		defer part.Close()
		reader := csv.NewReader(part)
		reader.TrimLeadingSpace = true
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		fi := -1
		li := -1
		ei := -1
		pi := -1
		di := -1
		fn := ""
		ln := ""
		ea := ""
		ps := ""
		ds := ""
		for i, v := range record {
			switch {
			case v == "First Name" || v[3:len(v)] == "First Name":
				fi = i
			case v == "Last Name":
				li = i
			case v == "Email":
				ei = i
			case v == "Position":
				pi = i
			case v == "Department":
				di = i
			}
		}
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if fi != -1 {
				fn = record[fi]
			}
			if li != -1 {
				ln = record[li]
			}
			if ei != -1 {
				ea = record[ei]
			}
			if pi != -1 {
				ps = record[pi]
			}
			if di != -1 {
				ds = record[di]
			}
			t := models.Target{
				FirstName:  fn,
				LastName:   ln,
				Email:      ea,
				Position:   ps,
				Department: ds,
			}
			ts = append(ts, t)
		}
	}
	return ts, nil
}

// CheckAndCreateSSL is a helper to setup self-signed certificates for the administrative interface.
func CheckAndCreateSSL(cp string, kp string) error {
	// Check whether there is an existing SSL certificate and/or key, and if so, abort execution of this function
	if _, err := os.Stat(cp); !os.IsNotExist(err) {
		return nil
	}
	if _, err := os.Stat(kp); !os.IsNotExist(err) {
		return nil
	}

	Logger.Printf("Creating new self-signed certificates for administration interface...\n")

	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	notBefore := time.Now()
	// Generate a certificate that lasts for 10 years
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		return errors.New(fmt.Sprintf("TLS Certificate Generation: Failed to generate a random serial number: %s", err))
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Gophish"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return errors.New(fmt.Sprintf("TLS Certificate Generation: Failed to create certificate: %s", err))
	}

	certOut, err := os.Create(cp)
	if err != nil {
		return errors.New(fmt.Sprintf("TLS Certificate Generation: Failed to open %s for writing: %s", cp, err))
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(kp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.New(fmt.Sprintf("TLS Certificate Generation: Failed to open %s for writing", kp))
	}

	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return errors.New(fmt.Sprintf("TLS Certificate Generation: Unable to marshal ECDSA private key: %v", err))
	}

	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keyOut.Close()

	Logger.Println("TLS Certificate Generation complete")
	return nil
}
