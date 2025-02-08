package main

import (
    "crypto/rand"
    "fmt"
    "regexp"
)

var ticketIDRegex = regexp.MustCompile(`\[Ticket:(\w+)\]`)

func extractTicketID(subject string) string {
    matches := ticketIDRegex.FindStringSubmatch(subject)
    if len(matches) > 1 {
        return matches[1]
    }
    return generateTicketID()
}

func generateTicketID() string {
    b := make([]byte, 6)
    rand.Read(b)
    return fmt.Sprintf("T%X", b)
}
