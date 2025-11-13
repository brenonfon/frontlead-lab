package main

import (
    "database/sql"
    "encoding/json"
    "encoding/xml"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"

    _ "github.com/lib/pq"
)

func getSalesforceConsumerKey() string {
    key := os.Getenv("SF_CONSUMER_KEY")
    if key == "" {
        log.Fatal("SF_CONSUMER_KEY environment variable is required")
    }
    return key
}

func getSalesforceConsumerSecret() string {
    secret := os.Getenv("SF_CONSUMER_SECRET")
    if secret == "" {
        log.Fatal("SF_CONSUMER_SECRET environment variable is required")
    }
    return secret
}

func getHubspotAPIKey() string {
    key := os.Getenv("HS_API_KEY")
    if key == "" {
        log.Fatal("HS_API_KEY environment variable is required")
    }
    return key
}

func getConnStr() string {
    host := os.Getenv("DB_HOST")
    if host == "" {
        host = "localhost" // Default for local development
    }
    
    port := os.Getenv("DB_PORT")
    if port == "" {
        port = "5432"
    }
    
    user := os.Getenv("DB_USER")
    if user == "" {
        user = "postgres"
    }
    
    password := os.Getenv("DB_PASSWORD")
    if password == "" {
        password = "postgres"
    }
    
    dbname := os.Getenv("DB_NAME")
    if dbname == "" {
        dbname = "salesforce"
    }
    
    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)
}

type SOAPEnvelope struct {
    XMLName xml.Name `xml:"Envelope"`
    Body    struct {
        XMLName       xml.Name `xml:"Body"`
        Notifications struct {
            XMLName      xml.Name `xml:"notifications"`
            Notification struct {
                XMLName xml.Name `xml:"Notification"`
                SObject struct {
                    ID        string `xml:"Id"`
                    FirstName string `xml:"FirstName"`
                    LastName  string `xml:"LastName"`
                    Email     string `xml:"Email"`
                } `xml:"sObject"`
            } `xml:"Notification"`
        } `xml:"notifications"`
    } `xml:"Body"`
}

func main() {
    mux := http.NewServeMux()
    
    mux.HandleFunc("/soap", soapHandler)
	mux.HandleFunc("/leads", leadsHandler)
    
    fmt.Println("Server is running on http://localhost:8081")
    fmt.Println("Routes:")
    fmt.Println("  GET /leads?email=<email> - Check if lead exists in Salesforce")
    fmt.Println("  POST /soap - Salesforce SOAP notifications listener")
    
    log.Fatal(http.ListenAndServe(":8081", mux))
}

func leadsHandler(w http.ResponseWriter, r *http.Request) {
    email := r.URL.Query().Get("email")
    isExistingLeadInSalesforce, err := isExistingLeadInSalesforce(email)

    if err != nil {
        http.Error(w, "Error checking lead", http.StatusInternalServerError)
        return
    }

    w.Header().Add("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{
        "isExistingLeadInSalesforce": isExistingLeadInSalesforce,
    })
}

func soapHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
        return
    }

    bodyBytes, err := io.ReadAll(r.Body)
    if err != nil {
        log.Printf("Error reading body: %v", err)
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    var env SOAPEnvelope
    if err := xml.Unmarshal(bodyBytes, &env); err != nil {
        log.Printf("Error parsing SOAP: %v", err)
        http.Error(w, "Invalid SOAP", http.StatusBadRequest)
        return
    }

    sobj := env.Body.Notifications.Notification.SObject
    fmt.Printf("📨 Received Contact: ID=%s, FirstName=%s, LastName=%s, Email=%s\n",
        sobj.ID, sobj.FirstName, sobj.LastName, sobj.Email)

    if sobj.ID != "" {
        if err := insertContact(sobj.ID, sobj.FirstName, sobj.LastName, sobj.Email); err != nil {
            log.Printf("❌ DB insert error: %v", err)
        } else {
            fmt.Printf("✅ Saved contact %s to Postgres\n", sobj.Email)
        }
    }

    w.Header().Set("Content-Type", "text/xml")
    fmt.Fprint(w, `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>
        <notificationsResponse xmlns="http://soap.sforce.com/2005/09/outbound">
            <Ack>true</Ack>
        </notificationsResponse>
    </soap:Body>
</soap:Envelope>`)
}

func insertContact(id, firstName, lastName, email string) error {
    db, err := sql.Open("postgres", getConnStr())
    if err != nil {
        return fmt.Errorf("DB connection error: %v", err)
    }
    defer db.Close()

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS contacts (
            sf_id TEXT PRIMARY KEY,
            first_name TEXT,
            last_name TEXT,
            email TEXT
        )
    `)
    if err != nil {
        return fmt.Errorf("Error creating table: %v", err)
    }

    _, err = db.Exec(`
        INSERT INTO contacts (sf_id, first_name, last_name, email)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (sf_id) DO UPDATE
        SET first_name = EXCLUDED.first_name,
            last_name = EXCLUDED.last_name,
            email = EXCLUDED.email
    `, id, firstName, lastName, email)
    if err != nil {
        return fmt.Errorf("Error inserting record: %v", err)
    }

    return nil
}

func isExistingLeadInSalesforce(email string) (bool, error) {
    db, err := sql.Open("postgres", getConnStr())
    if err != nil {
        return false, fmt.Errorf("DB connection error: %v", err)
    }
    defer db.Close()

    var count int
    query := "SELECT COUNT(*) FROM contacts WHERE email = $1"
    err = db.QueryRow(query, email).Scan(&count)
    if err != nil {
        return false, fmt.Errorf("Error querying database: %v", err)
    }

    return count > 0, nil
}

// func isExistingLeadInHubspot(email string) (bool, error) {
//     client := &http.Client{}
    
//     req, err := http.NewRequest("GET", "https://api.hubapi.com/crm/v3/objects/contacts", nil)
//     if err != nil {
//         return false, err
//     }

//    req.Header.Add("Authorization", "Bearer "+os.Getenv("HS_API_KEY"))
//    req.Header.Add("Content-Type", "application/json")

//    q := req.URL.Query()
//    q.Add("properties", "email")
//    q.Add("limit", "1")
//     req.URL.RawQuery = q.Encode()
    
//     // Make the request
//     resp, err := client.Do(req)
//     if err != nil {
//         return false, err
//     }
//     defer resp.Body.Close()

//     if resp.StatusCode != http.StatusOK {
//         return false, fmt.Errorf("HubSpot API returned status code: %d", resp.StatusCode)
//     }

//     // Parse HubSpot response
//     var hubspotResponse struct {
//         Results []struct {
//             Properties struct {
//                 Email string `json:"email"`
//             } `json:"properties"`
//         } `json:"results"`
//         Total int `json:"total"`
//     }

//     if err := json.NewDecoder(resp.Body).Decode(&hubspotResponse); err != nil {
//         return false, err
//     }

//     // Check if any contact matches the email
//     for _, contact := range hubspotResponse.Results {
//         if contact.Properties.Email == email {
//             return true, nil
//         }
//     }
    
//     return false, nil
// }

