package mcp

// Server defines the placeholder seam for the official Model Context Protocol Go SDK.
// Implementation will wrap github.com/modelcontextprotocol/go-sdk behind this seam.
type Server struct {
	Name    string
	Version string
}

func NewServer(name, version string) *Server {
	return &Server{
		Name:    name,
		Version: version,
	}
}
