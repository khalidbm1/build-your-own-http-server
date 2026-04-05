# Build Your Own HTTP Server

> "Why use `net/http` when you can mass-produce your own bugs from scratch?" — Ancient Gopher Proverb

## What Is This?

This is an HTTP server built from **raw TCP sockets** in Go. 
- No `net/http`. 
- No frameworks. 
- No safety net. 

### Just a developer, a `net.Listener`, 
### and a mass of `\r\n` strings that somehow speak HTTP.

# How Did I Do It? I went from "what even is a TCP connection?" 
# to 

# "oh no, I'm manually parsing headers at 2 AM" in about 9 stages of questionable life decisions.

## Features (that actually work, believe it or not)

- **Raw TCP Socket Handling** — because importing `net/http` felt like cheating
- **HTTP Request Parser** — splits strings on `\r\n` and prays
- **HTTP Response Builder** — assembles bytes into something browsers won't scream at
- **URL Router** — if/else statements with extra steps
- **Static File Serving** — serves files AND blocks hackers trying `../../etc/passwd` (you're welcome)
- **Concurrent Connections** — goroutines go brrrrr (up to 100 at once)
- **Timeouts** — so one slow client doesn't hold the entire server hostage
- **Panic Recovery** — the server catches its own crashes and sends a polite 500 instead of dying dramatically
- **CLI Flags** — configurable like a real server, because we have standards (low ones, but still)

## The 9 Stages of Grief... I Mean Development

| Stage | What We Did | Emotional State |
|-------|------------|-----------------|
| 1 | TCP Listener | "This is easy!" |
| 2 | Request Parser | "Wait, HTTP has rules?" |
| 3 | Response Builder | "Why does everything need `\r\n`?" |
| 4 | Router | "I am basically building Express.js" |
| 5 | Static Files | "Path traversal attacks are real?!" |
| 6 | Concurrency | "Goroutines are magic" |
| 7 | Error Handling | "Everything that can go wrong, will" |
| 8 | Timeouts | "Slow clients are the enemy" |
| 9 | Final Assembly | "It works. I don't know why, but it works." |

## Project Structure

```
.
├── cmd/server/main.go          # Where it all begins (and sometimes ends)
├── internal/
│   ├── server/server.go        # The beating heart (TCP listener)
│   ├── server/errors.go        # A catalog of everything that can go wrong
│   ├── request/request.go      # String splitting: the professional edition
│   ├── response/response.go    # Assembling bytes like IKEA furniture
│   ├── response/files.go       # File serving with trust issues
│   └── router/router.go        # Traffic control for HTTP requests
└── static/                     # HTML, CSS, JS — the holy trinity
```

## Quick Start

```bash
# Build it
go build -o httpserver ./cmd/server

# Run it
./httpserver -port 8080

# Marvel at your creation
curl http://localhost:8080/
curl http://localhost:8080/health
curl http://localhost:8080/api/info
```

## CLI Flags

```
-port          Port number (default: 8080)
-addr          Bind address (default: all interfaces)
-max-conns     Max concurrent connections (default: 100)
-read-timeout  Read timeout (default: 10s)
-write-timeout Write timeout (default: 10s)
-static        Static files directory (default: ./static)
```

## Endpoints

| Route | What It Does |
|-------|-------------|
| `/` | Serves a beautiful landing page |
| `/health` | Returns 200 OK (proof of life) |
| `/api/info` | Server info in JSON (it's basically an API now) |
| `/static/*` | Serves static files (securely!) |
| Everything else | 404, obviously |

## Security Features

- **Path traversal protection** — nice try with your `../../../etc/passwd`
- **Request size limits** — 8KB headers, 1MB body, no exceptions
- **Method validation** — only the HTTP methods we agreed on
- **Timeout protection** — slowloris attacks get the boot
- **Panic recovery** — the server stays alive even when handlers don't

## Things I Learned

1. HTTP is just text over TCP. Fancy, structured text, but still text.
2. `\r\n` is the most important string in web development.
3. Concurrency in Go is delightful until it isn't.
4. Every line of `net/http` source code was written by someone who suffered before you.
5. Building things from scratch is the best way to truly understand them.

## Could I Have Just Used `net/http`?

Yes. Absolutely. In a fraction of the time. With better performance, security, and reliability.

But where's the fun in that?

## License

Do whatever you want with this. If it breaks your production server, that's between you and your SRE team.

---

*Built with mass amounts of curiosity, Go 1.26, and zero external dependencies.*
