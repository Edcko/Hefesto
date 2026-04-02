---
name: go-testing
description: Go testing patterns with standard library, table-driven tests, and Bubbletea TUI testing
trigger: When writing Go tests, testing Go code, or Bubbletea TUI testing
version: 1.0.0
---

## When to Use

- Writing unit tests for Go packages
- Creating table-driven tests
- Testing HTTP handlers
- Testing Bubbletea TUI applications
- Benchmarking Go code

## Standard Go Testing

### Running Tests

```bash
go test ./...                    # All packages
go test ./pkg/service            # Specific package
go test -v ./...                 # Verbose output
go test -run TestName ./...      # Run specific test
go test -count=1 ./...           # Disable cache
go test -race ./...              # Race detector
```

### Table-Driven Tests (The Go Way)

```go
func TestParse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int
        wantErr bool
    }{
        {"valid number", "42", 42, false},
        {"negative", "-10", -10, false},
        {"invalid", "abc", 0, true},
        {"empty", "", 0, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Parse(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Parse() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Helpers

```go
// Helper function - marks itself as helper
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper() // Error reports caller's line, not this function
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("setup db: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}

// Cleanup pattern
func TestWithCleanup(t *testing.T) {
    file := createTempFile(t)
    t.Cleanup(func() { os.Remove(file.Name()) })
    // test continues...
}
```

### TestMain Setup/Teardown

```go
func TestMain(m *testing.M) {
    // Setup
    os.Setenv("TEST_MODE", "1")
    
    code := m.Run()
    
    // Teardown
    os.Unsetenv("TEST_MODE")
    os.Exit(code)
}
```

### Coverage

```bash
go test -cover ./...                     # Coverage summary
go test -coverprofile=coverage.out ./... # Profile file
go tool cover -func=coverage.out         # Function breakdown
go tool cover -html=coverage.out         # HTML report
```

### Build Tags

```go
//go:build integration
// +build integration

func TestIntegration(t *testing.T) {
    // Run with: go test -tags=integration ./...
}
```

### Benchmarks

```go
func BenchmarkParse(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Parse("42")
    }
}

// Run: go test -bench=. -benchmem
```

### Fuzzing (Go 1.18+)

```go
func FuzzParse(f *testing.F) {
    f.Add("42")  // Seed corpus
    f.Fuzz(func(t *testing.T, input string) {
        Parse(input) // Should not panic
    })
}

// Run: go test -fuzz=FuzzParse
```

## Mocking & Interfaces

### Interface-Based Mocking (No Framework)

```go
// Define interface for external dependency
type UserRepository interface {
    Find(id int) (*User, error)
    Save(user *User) error
}

// Real implementation
type DBUserRepo struct { db *sql.DB }

// Test mock
type MockUserRepo struct {
    users map[int]*User
    err   error
}

func (m *MockUserRepo) Find(id int) (*User, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.users[id], nil
}
```

### httptest for HTTP Handlers

```go
func TestHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/users/1", nil)
    rec := httptest.NewRecorder()
    
    handler := UserHandler{Repo: &MockUserRepo{}}
    handler.ServeHTTP(rec, req)
    
    if rec.Code != http.StatusOK {
        t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
    }
}
```

### testify Assertions (If Project Uses It)

```go
import "github.com/stretchr/testify/assert"

func TestWithTestify(t *testing.T) {
    got, err := Parse("42")
    
    assert.NoError(t, err)
    assert.Equal(t, 42, got)
    assert.Contains(t, []string{"a", "b"}, "a")
}
```

## Bubbletea TUI Testing

### Basic Program Test

```go
import (
    "github.com/charmbracelet/bubbletea"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/x/exp/teatest"
)

func TestModel(t *testing.T) {
    m := NewModel()
    tm := teatest.NewTestModel(t, m)
    defer tm.Quit()
    
    // Send key
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
    
    // Wait for final model
    teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
        return bytes.Contains(b, []byte("expected output"))
    })
}
```

### Testing Update Function

```go
func TestUpdate(t *testing.T) {
    m := NewModel()
    
    // Test key message
    msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
    newModel, cmd := m.Update(msg)
    
    // Check state change
    if newModel.(Model).cursor != 1 {
        t.Error("cursor should move")
    }
    
    // Check command
    if cmd != nil {
        t.Log("Command returned")
    }
}
```

### Testing View Output

```go
func TestView(t *testing.T) {
    m := NewModel()
    view := m.View()
    
    if !strings.Contains(view, "Welcome") {
        t.Error("view should contain welcome message")
    }
}
```

### Sending Messages

```go
// Key press
tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

// Custom message
tm.Send(customMsg{data: "test"})

// Window resize
tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
```

## Conventions

| Convention | Pattern |
|------------|---------|
| Test files | `*_test.go` in same package |
| Test names | `TestFunctionName`, `TestFunctionName_Scenario` |
| Table tests | `[]struct{name, input, want, wantErr}` |
| Helpers | Call `t.Helper()` at start |
| Parallel | Call `t.Parallel()` at start |
| Short tests | `if testing.Short() { t.Skip() }` |

### Skip Patterns

```go
func TestSlow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping slow test")
    }
    // Run with: go test -short ./...
}

func TestRequiresDocker(t *testing.T) {
    if _, err := exec.LookPath("docker"); err != nil {
        t.Skip("docker not installed")
    }
}
```
