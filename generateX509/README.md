# 为TLS服务器生成自签名X.509证书



### 一、仅仅生成并使用

```go
func main() {

	_gx := &generateX509.Info{}
	pk, err := _gx.Generate()
	if err != nil {
		log.Fatalln(err)
	}

	s := &http.Server{
		Addr:    "127.0.0.1:1443",
		Handler: nil,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*pk},
		},
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World!\r\n"))
	})
	log.Println(s.ListenAndServeTLS("", ""))
}
```



### 二、生成公私钥文件并使用

```go
func main() {

	_gx := &generateX509.Info{}
	pk, err := _gx.Generate()
	if err != nil {
		log.Fatalln(err)
	}

	crt, key, err := _gx.Pem(pk)
	if err != nil {
		log.Fatalln(err)
	}

	crtF, err := os.Create("/tmp/localhost.crt")
	if err != nil {
		log.Fatalln(err)
	}
	defer crtF.Close()
	keyF, err := os.Create("/tmp/localhost.key")
	if err != nil {
		log.Fatalln(err)
	}
	defer keyF.Close()

	_, _ = crtF.Write(crt)
	_, _ = keyF.Write(key)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World!\r\n"))
	})
	log.Println(http.ListenAndServeTLS(":1443", "/tmp/localhost.crt", "/tmp/localhost.key", nil))
}
```


