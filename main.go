package main

import (
	"bufio"
	"bytes"
	"encoding/base32"
	"fmt"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func display(key *otp.Key, data []byte) {
	fmt.Printf("Issuer:       %s\n", key.Issuer())
	fmt.Printf("Account Name: %s\n", key.AccountName())
	fmt.Printf("Secret:       %s\n", key.Secret())
	fmt.Println("Writing PNG to qr-code.png....")
	ioutil.WriteFile("qr-code.png", data, 0644)
	fmt.Println("")
	fmt.Println("Please add your TOTP to your OTP Application now!")
	fmt.Println("")
}

func promptForPasscode() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Passcode: ")
	text, _ := reader.ReadString('\n')
	return text
}

// Demo function, not used in main
// Generates Passcode using a UTF-8 (not base32) secret and custom paramters
func GeneratePassCode(utf8string string) string {
	secret := base32.StdEncoding.EncodeToString([]byte(utf8string))
	passcode, err := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA512,
	})
	if err != nil {
		panic(err)
	}
	return passcode
}

func main() {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Example.com",
		AccountName: "alice@example.com",
	})
	if err != nil {
		panic(err)
	}
	// Convert TOTP key into a PNG
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		panic(err)
	}
	png.Encode(&buf, img)

	// display the QR code to the user.
	display(key, buf.Bytes())

	validHandler := func(passcode string) string {
		valid := totp.Validate(passcode, key.Secret())
		if valid {
			return "valid"
		} else {
			return "invalid"
		}
	}

	// Hello world, the web server
	helloHandler := func(w http.ResponseWriter, r *http.Request) {
		// performing external query for working ssh connection
		passcode := r.URL.Query().Get("passcode")
		out := validHandler(passcode)
		_ = err
		io.WriteString(w, string(out))
	}

	http.HandleFunc("/totp", helloHandler)

	log.Fatal(http.ListenAndServe(":8888", nil))

	//// Now Validate that the user's successfully added the passcode.
	//fmt.Println("Validating TOTP...")
	//passcode := promptForPasscode()
	//valid := totp.Validate(passcode, key.Secret())
	//if valid {
	//	println("Valid passcode!")
	//	os.Exit(0)
	//} else {
	//	println("Invalid passcode!")
	//	os.Exit(1)
	//}
}