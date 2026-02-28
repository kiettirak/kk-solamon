# solarman-go-client

A Go client for fetching data from the [Solarman API](https://doc.solarmanpv.com/) (globalapi.solarmanpv.com).

## Features

- Token-based authentication with auto-refresh
- Fetch station (plant) list
- Fetch real-time device data
- Save results as timestamped JSON files

## Project Structure

```
solarman-go-client/
├── cmd/solarman-client/
│   └── main.go              # Entry point
├── internal/
│   ├── config/config.go     # Load .env configuration
│   └── solarman/
│       ├── auth.go          # Authentication & token management
│       ├── client.go        # HTTP client for Solarman API
│       ├── endpoints.go     # API endpoint constants
│       └── models.go        # Request/Response structs
└── internal/service/
    └── data_service.go      # Orchestrates API calls & saves JSON
```

## Requirements

- [Go 1.21+](https://go.dev/dl/)
- Solarman account at [home.solarmanpv.com](https://home.solarmanpv.com)
- Solarman API credentials (App ID & App Secret) from [Solarman Developer Portal](https://doc.solarmanpv.com/)

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/kiettirak/solarman-go-client.git
cd solarman-go-client
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Configure environment variables

Copy `.env.example` to `.env` and fill in your credentials:

```bash
cp .env.example .env
```

Edit `.env`:

```ini
# Solarman API Credentials (from developer portal)
API_ID=your_app_id
API_SECRET=your_app_secret
BASE_URL=https://globalapi.solarmanpv.com

# Solarman Account (same as home.solarmanpv.com login)
EMAIL=your_email@example.com
PASSWORD=your_password

# Optional
LOG_LEVEL=info
OUTPUT_DIR=./output
```

> **Note:** Password with special characters (e.g. `#`) must be quoted: `PASSWORD="your#pass"`

### 4. Run

```bash
go run ./cmd/solarman-client/main.go
```

## Output

JSON files are saved to `./output/` with timestamps:

```
output/
└── 20260228_220132_stations.json
```

## Authentication Notes

| Field | Format |
|-------|--------|
| `appSecret` | plain text |
| `password` | SHA-256 hash (lowercase hex) |

Token is valid for **60 days** and auto-refreshed on expiry.

## License

MIT
## Setup Instructions
1. **Clone the repository:**
   ```
   git clone <repository-url>
   cd solarman-go-client
   ```

2. **Install dependencies:**
   Ensure you have Go installed, then run:
   ```
   go mod tidy
   ```

3. **Configure environment variables:**
   Copy `.env.example` to `.env` and fill in the required values for your Solarman API credentials.

4. **Run the application:**
   ```
   go run cmd/solarman-client/main.go
   ```

## Usage
After setting up the project, you can use the Solarman Go Client to retrieve data from the Solarman API. The application will handle authentication and make requests to the appropriate endpoints.

## Contributing
Contributions are welcome! Please submit a pull request or open an issue for any enhancements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.