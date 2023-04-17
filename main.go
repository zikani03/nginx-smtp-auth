package main

import (
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"github.com/alexedwards/scs/v2"
)

var (
	listenAddr          string
	smtpHost            string
	smtpPort            int
	smtpEnableTLS       bool
	smtpSkipVerifyCerts bool
)

var sessionManager *scs.SessionManager

func init() {
	flag.StringVar(&listenAddr, "listen", ":9000", "listen address for the server")
	flag.StringVar(&smtpHost, "smtp-host", "smtp.example.com", "SMTP host")
	flag.IntVar(&smtpPort, "smtp-port", 465, "SMTP Port")
	flag.BoolVar(&smtpEnableTLS, "smtp-tls-enabled", true, "Use TLS/SSL connection")
	flag.BoolVar(&smtpSkipVerifyCerts, "skip-verification", true, "skip verification of invalid ssl/tls certs")
}

func main() {
	flag.Parse()

	sessionManager = scs.New()
	
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello there :")
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, req *http.Request) {
		authHeader, ok := req.Header["Authorization"]
		if !ok || len(authHeader) != 1 {
			w.Header().Add("WWW-Authenticate", "Basic")
			http.Error(w, "failed to decode auth headr", http.StatusUnauthorized)
			return
		}

		decodedData, err := base64.StdEncoding.DecodeString(strings.Replace(authHeader[0], "Basic ", "", 1))

		if err != nil {
			fmt.Fprintf(os.Stdout, "failed to decode from %s got error:%s", authHeader[0], err)
			http.Error(w, "failed to decode auth headr", http.StatusUnauthorized)
			return
		}
		decoded := string(decodedData)

		if decoded == "" {
			http.Error(w, "failed to decode auth headr", http.StatusUnauthorized)
			return
		}
		authHeaderArr := strings.SplitN(decoded, ":", 2)
		user := authHeaderArr[0]
		passwd := authHeaderArr[1]
		servername := fmt.Sprintf("%s:%d", smtpHost, smtpPort)

		host, _, _ := net.SplitHostPort(servername)
		tlsconfig := &tls.Config{
			InsecureSkipVerify: smtpSkipVerifyCerts,
			ServerName:         host,
		}

		conn, err := tls.Dial("tcp", servername, tlsconfig)

		if err != nil {
			http.Error(w, "failed to connect to smtp host", http.StatusUnauthorized)
			return
		}

		fmt.Println("dialing the smtp server at ", servername)
		c, err := smtp.NewClient(conn, host)
		if err != nil {
			http.Error(w, "failed to connect to smtp host", http.StatusUnauthorized)
			return
		}
		defer c.Quit()

		err = c.Auth(smtp.PlainAuth("", user, passwd, host))
		if err != nil {
			fmt.Fprintf(os.Stdout, "failed to authenticate to %s as user %s got error:%s", servername, user, err)
			http.Error(w, "invalid username or password", http.StatusUnauthorized)
			return
		}
			
		err = sessionManager.RenewToken(req.Context())
		if err != nil {
			http.Error(w, "invalid username or password", http.StatusUnauthorized)
			return
		}

		sessionManager.Put(req.Context(), "auth.user", user)
		fmt.Fprintf(w, "authenticated successfully")
	})

	fmt.Printf("Starting server at %s\n", listenAddr)
	err := http.ListenAndServe(listenAddr, sessionManager.LoadAndSave(mux))
	if err != nil {
		log.Fatalf("failed to start server %v", err)
	}
}
