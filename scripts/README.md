# HubSpot Setup Scripts

## Setup HubSpot Properties

This script creates the custom contact properties needed for the application in your HubSpot account.

### Properties Created

1. **business_needs** - Customer's business needs and requirements (textarea)
2. **integration_timeline** - Expected timeline for integration (text)
3. **seats_or_extensions** - Number of seats or extensions needed (text)
4. **last_offer_summary** - Summary of the last offer provided (textarea)
5. **interest_topic** - Topic of interest for the contact (text)
6. **interest_source** - Source of the contact's interest (text)
7. **campaign_offer** - Campaign offer associated with the contact (text)

### How to Run

1. Make sure your `config/.env` file has the `HS_API_KEY` set:
   ```bash
   HS_API_KEY=your_hubspot_api_key_here
   ```

2. Run the setup script from the project root:
   ```bash
   go run scripts/setup_hubspot_properties.go
   ```

### Expected Output

```
🚀 HubSpot Property Setup Script
================================
📝 Found API Key: pat-...xxxx
📋 Creating 7 custom properties in HubSpot...

[1/7] Creating property: business_needs
✅ Created property: Business Needs (business_needs)
[2/7] Creating property: integration_timeline
✅ Created property: Integration Timeline (integration_timeline)
...

================================
✅ Successfully created: 7 properties
🎉 Setup complete!
```

### Notes

- If a property already exists, the script will skip it (you'll see ⚠️ warning)
- The script uses the same API key from your `.env` file
- All properties are created in the "Contact Information" group
- You only need to run this script once per HubSpot account
