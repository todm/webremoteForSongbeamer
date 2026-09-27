package main

import (
	"context"
	"fmt"
	"image/jpeg"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/todm/webremoteForSongbeamer/thumbnailstealer"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type staticFS struct {
	root http.FileSystem
}

var busy = sync.Mutex{}

func main() {
	defer exit()

	err := ReadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	router := http.NewServeMux()

	router.HandleFunc("/osc", handleWebsocket)
	router.HandleFunc("/thumbnails", handleThumbnails)
	router.HandleFunc("/sbfiles/playlist", handleOSCFile("/playlist/filename"))
	router.HandleFunc("/sbfiles/presentation", handleOSCFile("/presentation/filename"))
	router.HandleFunc("/sbfiles/video", handleOSCFile("/video/filename"))
	router.Handle("/", http.FileServer(
		&staticFS{
			root: http.Dir(ResolvePath(cfg.StaticDir)),
		},
	))

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	fmt.Printf("Starting server on port %d\n", cfg.Port)

	println("\nAvailable at:")
	addresses, _ := GetIpAddresses()
	for _, addr := range addresses {
		var protocol string
		if cfg.UseTLS {
			protocol = "https"
		} else {
			protocol = "http"
		}
		fmt.Printf("    %s://%s:%d\n", protocol, addr, cfg.Port)
	}
	print("\n")

	if cfg.UseTLS {
		err = server.ListenAndServeTLS(ResolvePath(cfg.CertFile), ResolvePath(cfg.KeyFile))
	} else {
		err = server.ListenAndServe()
	}

	if err != nil {
		fmt.Println(err)
	}
}

func (fs *staticFS) Open(name string) (http.File, error) {
	file, err := fs.root.Open(name)
	if os.IsNotExist(err) {
		return fs.root.Open("index.html")
	}
	return file, err
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	udp, err := net.Dial("udp", cfg.SongbeamerAddr)
	if err != nil {
		return
	}
	defer udp.Close()

	fmt.Println("New connection ", r.RemoteAddr)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				cancel()
				return
			}

			busy.Lock()
			udp.Write(message)
			busy.Unlock()
		}

	}()

	go func() {
		buf := make([]byte, 2048)
		for {
			n, err := udp.Read(buf)
			if err != nil {
				cancel()
				return
			}
			ws.WriteMessage(websocket.BinaryMessage, buf[:n])
		}
	}()

	<-ctx.Done()
	udp.Close()
	ws.Close()
	fmt.Println("Closing connection ", r.RemoteAddr, err)
}

func handleOSCFile(address string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := &OSCMessage{
			Address: address,
		}

		res, err := RequestOSC(cfg.SongbeamerAddr, *msg)
		if err != nil || len(res.Arguments) == 0 {
			http.Error(w, "Could not get osc response", 500)
			return
		}

		if filePath, ok := res.Arguments[0].(string); ok {
			w.Header().Set("X-OSC-FILEPATH", filePath)
			http.ServeFile(w, r, filePath)
			return
		}
		http.Error(w, "Could not read file", 500)
	}
}

func handleThumbnails(w http.ResponseWriter, r *http.Request) {
	busy.Lock()
	defer busy.Unlock()

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "no-store")

	numThumbnails := 0
	columns := 0

	if c := r.URL.Query().Get("cols"); c != "" {
		fmt.Sscanf(c, "%d", &columns)
		msg := &OSCMessage{
			Address: "/presentation/pagecount",
		}
		res, err := RequestOSC(cfg.SongbeamerAddr, *msg)
		if err != nil || len(res.Arguments) == 0 {
			http.Error(w, "Could not get osc response", 500)
			return
		}

		if val, ok := res.Arguments[0].(int32); ok {
			numThumbnails = int(val)
		}
	}

	thumbnails, err := thumbnailstealer.Steal(cfg.ThumbnailSkipReset, numThumbnails, columns)

	if err != nil {
		http.Error(w, "failed getting thumbnails", http.StatusInternalServerError)
		return
	}

	err = jpeg.Encode(w, thumbnails, &jpeg.Options{
		Quality: cfg.ThumbnailQuality,
	})

	if err != nil {
		http.Error(w, "failed to encode image", http.StatusInternalServerError)
		return
	}
}

func GetIpAddresses() ([]string, error) {
	var addresses []string
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				addresses = append(addresses, ip4.String())
			}
			// Uncomment this if you want to include IPv6 addresses as well
			// else {
			// 	addresses = append(addresses, fmt.Sprintf("[%s]", ip))
			// }
		}
	}

	return addresses, nil
}

func exit() {
	fmt.Println("Press Enter to exit...")
	fmt.Scanf("%s")
}
