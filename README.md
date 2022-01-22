# gmail

library for sending email using gmail credentials.



## Usage

```go
package main

import (
    ...
    
    "github.com/t4ke0/gmail"
    ...
)

func main() {

    username := os.Getenv("GMAIL_USERNAME")
    password := os.Getenv("GMAIL_PASSWORD")

    em := gmail.NewEmail(username, password, gmail.EmailConfig{
        From:        username,
        To:          []string{""}, // put your recipients here
        Subject:     "", // subject here
        MessageText: "", // text body message
        Attachements: []string{}, // files attachments here.
    })

    if err := em.Marshal().Send().Error(); err != nil {
        log.Fatal(err)
    }
}
```
