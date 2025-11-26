# Salesforce OAuth 2.0 Web Server Flow Setup

## Configure Your Salesforce Connected App

1. **In Salesforce Setup:**
   - Go to: Setup → App Manager → New Connected App

2. **Basic Information:**
   - Connected App Name: `Your App Name`
   - API Name: `Your_App_Name`
   - Contact Email: your email

3. **API (Enable OAuth Settings):**
   - ✅ Check "Enable OAuth Settings"
   - **Callback URL:** `http://localhost:8081/auth/salesforce/callback`
   - **Selected OAuth Scopes:**
     - Full access (full)
     - Perform requests at any time (refresh_token, offline_access)
     - Access and manage your data (api)
   
4. **Save and Continue**

5. **Get Your Credentials:**
   - Click "Manage Consumer Details"
   - Copy the **Consumer Key** and **Consumer Secret**

## Update Your .env File

```env
PORT=8081
SF_URL=https://login.salesforce.com
SF_CONSUMER_KEY=your_actual_consumer_key_here
SF_CONSUMER_SECRET=your_actual_consumer_secret_here
SF_REDIRECT_URI=http://localhost:8081/auth/salesforce/callback
```

**For Sandbox:** Use `SF_URL=https://test.salesforce.com`

## Run the Application

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Authenticate:**
   - The server will print: `📱 To authenticate, visit: http://localhost:8081/auth/salesforce`
   - Open that URL in your browser
   - You'll be redirected to Salesforce login
   - Log in and authorize the app
   - You'll be redirected back with a success message

3. **Use the API:**
   Once authenticated, you can use the Lead endpoints:
   
   - **Create Lead:**
     ```bash
     curl -X POST http://localhost:8081/leads \
       -H "Content-Type: application/json" \
       -d '{
         "FirstName": "John",
         "LastName": "Doe",
         "Company": "Acme Corp",
         "Email": "john@acme.com",
         "Status": "Open - Not Contacted"
       }'
     ```
   
   - **Get Lead:**
     ```bash
     curl http://localhost:8081/leads/{LEAD_ID}
     ```
   
   - **Update Lead:**
     ```bash
     curl -X PATCH http://localhost:8081/leads/{LEAD_ID} \
       -H "Content-Type: application/json" \
       -d '{
         "Status": "Working - Contacted"
       }'
     ```

## Troubleshooting

- **"redirect_uri_mismatch" error:** Make sure the redirect URI in your `.env` matches exactly what's configured in your Connected App
- **"invalid_client_id" error:** Double-check your Consumer Key
- **"invalid_client" error:** Verify your Consumer Secret is correct
