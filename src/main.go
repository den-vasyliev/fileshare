package main

import (
    "fmt"
    "net/url"
    "net/http"
    "log"
    "io/ioutil"
    "os"
    "strings"
            "crypto/hmac"
        "crypto/sha1"
        "encoding/base64"
        "time"
       "strconv"
       "errors"
       "github.com/go-redis/redis"
       "flag"
)


var (
    AppRole *string
    //AppPort app port
    AppPort *string
    // AppLicense app
    AppCache = getEnv("APP_CACHE", "127.0.0.1")
    // AppCachePort app
    AppCachePort = getEnv("APP_CACHE_PORT", "6379")
    // Environment app
    Environment = ""
    // CACHE redis conn
    CACHE *redis.Client
    // Role application name
    Role = ""
    // Version app
    Version = "1.0.12"
)


// sign URL with key and return sign
func sign(url, keyName string, expiration time.Time, s string) ([]byte, error) {
      
        keyPath := os.Getenv("KEY_PATH")
        // Note: consider using the GCP Secret Manager for managing access to your
        // signing key(s).
        key, err := readKeyFile(keyPath)
        if err != nil {
                log.Print(err)
        return nil, fmt.Errorf("failed to readKeyFile: %+v", err)
        }

        sep := "?"
        if strings.Contains(url, "?") {
                sep = "&"
        }
        url += sep
        url += fmt.Sprintf("Expires=%d", expiration.Unix())
        url += fmt.Sprintf("&KeyName=%s", keyName)
        log.Print(url)
        mac := hmac.New(sha1.New, key)
        mac.Write([]byte(url))
        sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
        //url += fmt.Sprintf("&Signature=%s", sig)
        

        if s == sig {

        return []byte(sig), nil
        }else{
            err = errors.New("not valid")
        return nil, fmt.Errorf(" sign: %+v", err)
    }
}



// readKeyFile reads the base64url-encoded key file and decodes it.
func readKeyFile(path string) ([]byte, error) {
        b, err := ioutil.ReadFile(path)
        if err != nil {
                return nil, fmt.Errorf("failed to read key file: %+v", err)
        }
        d := make([]byte, base64.URLEncoding.DecodedLen(len(b)))
        n, err := base64.URLEncoding.Decode(d, b)
        if err != nil {
                return nil, fmt.Errorf("failed to base64url decode: %+v", err)
        }
        return d[:n], nil
}

// start server
type indexHandler struct {
}

func (h *indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    m, err := url.ParseQuery(r.URL.RawQuery)
    if err != nil {
        log.Print("RawQuery")
        http.Error(w, http.StatusText(403), 403)
        return
        }
        
        req, err := fmt.Printf("Req: %s %s %s\n", r.URL.Path, r.Host, r.URL.RawQuery )

        // The path to a file containing the base64-encoded signing key
        if err != nil {
        log.Print(err)
        log.Print("Checklist")
        http.Error(w, http.StatusText(403), 403)
        return
        }else{
            log.Print(req)
        }

    if len(m["Expires"]) == 0 {
        log.Print("Expires")
        http.Error(w, http.StatusText(410), 410)
        return
    }

    if len(m["KeyName"]) == 0 {
        log.Print("KeyName")
        http.Error(w, http.StatusText(403), 403)
        return
    }

    ttl, err := strconv.Atoi(m["Expires"][0])

    if err != nil {
        log.Print("Atoi")
        log.Print(err)
        http.Error(w, http.StatusText(403), 403)
        return
    }
    
    scheme := "https://"
    if r.URL.Scheme == "" {
        scheme = "https://"
    }
    // validate and compare sign
    if len(m["Signature"]) == 0 {

        log.Print("Signature")

        http.Error(w, http.StatusText(403), 403)
        return
    }else{

    s, err := sign(fmt.Sprintf("%s%s%s", scheme, r.Host, r.URL.Path),m["KeyName"][0],time.Unix(int64(ttl), 0),m["Signature"][0])

    if err != nil {
        log.Print("Not Valid")
        log.Print(string(s),m["Signature"][0],err)
        http.Error(w, http.StatusText(403), 403)
    }else{
    log.Print(string(s),m["Signature"][0])
    
    // get mapping from data service
    
    cached, err := CACHE.Get(string(r.URL.Path)).Result()
    if err != nil {
        log.Print("No Map")
        //http.Error(w, http.StatusText(403), 403)
        cached = string(r.URL.Path)
    }else{
        log.Print(cached)
    }
        log.Print(cached)
    http.ServeFile(w, r, fmt.Sprintf("./fileshare/%s",cached))
      //  }
    }

    }
}

// getEnv get key environment variable if exist otherwise return defalutValue
func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return defaultValue
    }
    return value
}

// cache for mapping
func cache() {
    // Connect to cache
    CACHE = redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", AppCache, AppCachePort),
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    _, err := CACHE.Ping().Result()
    if err != nil {
        log.Print(err)
    }
}


// main
func main() {

    AppName := flag.String("name", "cdn-proxy", "application name")
    AppPort := flag.String("port", "8888", "application port")
    Environment = fmt.Sprintf("%s:%s", *AppName, Version)
    
    flag.Parse()
    
    log.Printf("[%s]: %s", Environment, *AppPort)
    cache()
    httpServer := &http.Server{Addr: fmt.Sprintf("%v:%v", "", *AppPort), Handler: &indexHandler{}}
        err := httpServer.ListenAndServe()
        if err != nil {
            fmt.Println("Err from server")
            fmt.Println(err)
        }

}