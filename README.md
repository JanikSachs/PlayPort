# PlayPort

PlayPort makes it easy to transfer your playlists between all major music platforms - fast, simple, and seamless.

## 🎵 Features

- **Easy Transfer**: Transfer playlists from one music platform to another in just a few clicks
- **Clean Architecture**: Modular design with extensible provider interface
- **Server-Side Rendering**: HTMX-powered frontend with no JavaScript framework overhead
- **Modern UI**: Beautiful, responsive interface using Bulma CSS
- **Provider System**: Extensible provider interface for adding new music platforms

## 🏗️ Architecture

PlayPort follows a clean architecture pattern with clear separation of concerns:

```
/cmd/playport/        -> Application entrypoint
/internal/server/     -> HTTP server setup and routing
/internal/handlers/   -> HTMX endpoints and HTML responses
/internal/providers/  -> Music platform integrations
/internal/models/     -> Domain models (Playlist, Track, Connection)
/internal/services/   -> Business logic for playlist transfers
/web/templates/       -> HTML templates (Go templates)
/web/static/css/      -> CSS styles
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- A modern web browser

### Running the Application

1. Clone the repository:
```bash
git clone https://github.com/JanikSachs/PlayPort.git
cd PlayPort
```

2. Build the application:
```bash
go build -o playport ./cmd/playport
```

3. Run the server:
```bash
./playport
```

4. Open your browser and navigate to:
```
http://localhost:8080
```

### Alternative: Run without building

```bash
go run ./cmd/playport
```

## 🧪 Testing

Run all tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

Run tests for a specific package:
```bash
go test ./internal/providers -v
go test ./internal/handlers -v
```

## 📖 Usage

### MVP Flow

The application currently includes a mock provider that demonstrates the complete workflow:

1. **Select Source Provider**: Navigate to the Transfer page and select "Mock Music" as your source
2. **Load Playlists**: Click "Load Playlists" to fetch all available playlists
3. **Choose Playlist**: Select a playlist you want to transfer
4. **Select Target Provider**: Choose "Mock Music" as the target provider
5. **Transfer**: Click "Transfer" and watch the progress in real-time
6. **Done!**: Your playlist has been transferred

### Adding New Providers

To add support for a new music platform:

1. Create a new provider in `internal/providers/` implementing the `Provider` interface:

```go
type Provider interface {
    Name() string
    Authenticate() error
    GetPlaylists() ([]models.Playlist, error)
    ExportPlaylist(id string) (models.Playlist, error)
    ImportPlaylist(p models.Playlist) error
}
```

2. Register the provider in `cmd/playport/main.go`:

```go
// Example for Spotify
spotifyProvider := providers.NewSpotifyProvider()
transferService.RegisterProvider(spotifyProvider)
```

## 🛠️ Technology Stack

- **Backend**: Go (Golang) with net/http
- **Frontend**: HTMX for dynamic interactions
- **CSS**: Bulma framework with custom overrides
- **Templates**: Go's html/template package
- **Architecture**: Clean architecture with provider pattern

## 📁 Project Structure

```
PlayPort/
├── cmd/
│   └── playport/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── handlers/
│   │   ├── handlers.go          # HTTP request handlers
│   │   └── handlers_test.go     # Handler tests
│   ├── models/
│   │   └── playlist.go          # Domain models
│   ├── providers/
│   │   ├── provider.go          # Provider interface
│   │   ├── mock.go              # Mock provider implementation
│   │   └── mock_test.go         # Provider tests
│   ├── server/
│   │   └── server.go            # HTTP server setup
│   └── services/
│       └── transfer.go          # Transfer business logic
├── web/
│   ├── static/
│   │   └── css/
│   │       └── custom.css       # Custom styles
│   └── templates/
│       ├── home.html            # Home page
│       ├── providers.html       # Providers page
│       ├── transfer.html        # Transfer page
│       ├── playlist-list.html   # Playlist list partial
│       └── transfer-result.html # Transfer result partial
├── go.mod
├── go.sum
└── README.md
```

## 🔌 Provider Interface

The `Provider` interface defines the contract for all music platform integrations:

```go
type Provider interface {
    // Name returns the provider's name (e.g., "Spotify", "Apple Music")
    Name() string

    // Authenticate authenticates with the provider's API
    Authenticate() error

    // GetPlaylists retrieves all playlists for the authenticated user
    GetPlaylists() ([]models.Playlist, error)

    // ExportPlaylist exports a specific playlist by ID
    ExportPlaylist(id string) (models.Playlist, error)

    // ImportPlaylist imports a playlist into the provider
    ImportPlaylist(p models.Playlist) error
}
```

## 🎨 Frontend Features

- **HTMX Integration**: Server-driven UI updates without page reloads
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **Real-time Progress**: Live transfer status updates
- **Smooth Animations**: CSS transitions for better UX
- **Clean UI**: Modern design with Bulma CSS framework

## 🧩 Extensibility

PlayPort is designed to be easily extended:

- **Add New Providers**: Implement the Provider interface for any music platform
- **Custom Transfer Logic**: Extend the TransferService for advanced features
- **Enhanced UI**: Add new templates and endpoints for additional features
- **API Integration**: Add RESTful API endpoints for programmatic access

## 🔐 Security Notes

- Never store sensitive credentials in the code
- Use environment variables for API keys and secrets
- Implement proper OAuth flows for real providers
- Validate and sanitize all user inputs
- Use HTTPS in production

## 📝 Future Enhancements

Potential features for future development:

- Real provider integrations (Spotify, Apple Music, YouTube Music, etc.)
- User authentication and session management
- Playlist transfer history
- Batch transfers
- Track matching algorithms (for cross-platform track resolution)
- Progress persistence and resume capability
- RESTful API
- Docker containerization
- CI/CD pipeline with GitHub Actions

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is open source and available under the MIT License.

## 🙏 Acknowledgments

- Built with Go and HTMX
- UI powered by Bulma CSS
- Inspired by the need for seamless playlist migration

