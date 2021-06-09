package cache

import (
    "crypto"
    "fmt"
    "os"
    "reflect"
)

type Cache struct {
    entries         map[string][]*CacheEntry
}

type CacheEntry struct {
    ServiceName         string
    Hash                string
}

func New() *Cache {
    entries := make(map[string][]*CacheEntry)

    return &Cache{
        entries: entries,
    }
}

func (cache *Cache) find(ozoneWorkingDir string, serviceName string) *CacheEntry {
    entries, ok := cache.entries[ozoneWorkingDir]

    if ok {
        for _, ce := range entries {
            if ce.ServiceName == serviceName {
                return ce
            }
        }
    }

    return nil
}

func (cache *Cache) Check(ozoneWorkingDir string, service string, ozoneFileAndDirHash string) bool {
    if ce := cache.find(ozoneWorkingDir, service); ce != nil {
        if ce.Hash == ozoneFileAndDirHash {
            return true
        }
    }
    return false
}



func remove(s []*CacheEntry, i int) []*CacheEntry {
    s[len(s)-1], s[i] = s[i], s[len(s)-1]
    return s[:len(s)-1]
}

func (cache *Cache) Update(ozoneWorkingDir string, service string, ozoneFileAndDirHash string) bool {
    if ce := cache.find(ozoneWorkingDir, service); ce != nil {
        ce.Hash = ozoneFileAndDirHash
        return true
    }
    cacheEntry := CacheEntry{
        ServiceName: service,
        Hash: ozoneFileAndDirHash,
    }
    cache.entries[ozoneWorkingDir] = append(cache.entries[ozoneWorkingDir], &cacheEntry)
    return true
}



func FileLastEdit(stringPath string) (int64, error) {
    file, err := os.Stat(stringPath)

    if err != nil {
        return 0, err
    }

    return file.ModTime().Unix(), nil
}

func Hash(objs ...interface{}) string {
    digester := crypto.MD5.New()
    for _, ob := range objs {
        fmt.Fprint(digester, reflect.TypeOf(ob))
        fmt.Fprint(digester, ob)
    }
    return fmt.Sprintf("%x", digester.Sum(nil))
}

