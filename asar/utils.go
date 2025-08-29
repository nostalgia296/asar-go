package asar

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"strconv"
	"time"
)

var (
filesToDump []string
ts = time.Now()
)

type fileEntry struct {
	Offset string `json:"offset"`
	Size   int    `json:"size"`
}

type directoryEntry struct {
	Files map[string]interface{} `json:"files"`
}

func makeJson(sourceFlag string) []byte {
	offset := 0

	index := directoryEntry{
		Files: make(map[string]interface{}),
	}

	walkRoot, err := filepath.EvalSymlinks(sourceFlag)
	if err != nil {
		fmt.Println("出现错误")
	}
	
	
	err = filepath.Walk(walkRoot, func(fullPath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fullPath == walkRoot {
			return nil
		}

		if info.IsDir() {

		} else if info.Mode().IsRegular() {
			p := strings.TrimPrefix(fullPath, walkRoot+"/")

			parts := strings.Split(p, "/")
			
			ptr := &index
			for _, part := range parts[:len(parts)-1] {
				if ptr.Files[part] == nil {
					ptr.Files[part] = directoryEntry{
						Files: make(map[string]interface{}),
					}
				}

				v, ok := ptr.Files[part].(directoryEntry)
				if ok {
					ptr = &v
				} else {
					return errors.New("unknonwn file type")
				}
			}

			ptr.Files[filepath.Base(p)] = fileEntry{
				Offset: fmt.Sprintf("%d", offset),
				Size:   int(info.Size()),
			}

			filesToDump = append(filesToDump, fullPath)
			offset += int(info.Size())
		}

		return nil
	})
	
	if err != nil {
	    fmt.Println("出现错误")
	}

	data, err := json.Marshal(index)
	if err != nil {
		fmt.Println("出现错误")
	}
	
	return data
	}
	
	
	
	
func Pack(sourceFlag, outputFlag string){
    data := makeJson(sourceFlag)
	fd, err := os.Create(outputFlag)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	paddingLength := (4 - len(data)%4) % 4



	preheader := make([]byte, 16)


	binary.LittleEndian.PutUint32(preheader[0:4], 4)
	binary.LittleEndian.PutUint32(preheader[4:8], uint32(len(data)+paddingLength+8))

	binary.LittleEndian.PutUint32(preheader[8:12], uint32(len(data)+paddingLength+4))
	binary.LittleEndian.PutUint32(preheader[12:16], uint32(len(data)))

	fd.Write(preheader)
	fd.Write(data)

	if paddingLength > 0 {
		fd.Write(make([]byte, paddingLength))
	}

	for _, file := range filesToDump {
		infd, err := os.Open(file)
		if err != nil {
			panic(err)
		}

		io.Copy(fd, infd)
		infd.Close()

	}
	}

func readu32(buffer []byte) uint32 {
	return binary.LittleEndian.Uint32(buffer[:4])
}
	
func readHeader(reader *os.File) (uint32, map[string]interface{}, error) {
	headerBuffer := make([]byte, 16)
	if _, err := reader.Read(headerBuffer); err != nil {
		return 0, nil, err
	}

	headerSize := readu32(headerBuffer[4:8])
	jsonSize := readu32(headerBuffer[12:16])

	jsonBuffer := make([]byte, jsonSize)
	if _, err := reader.Read(jsonBuffer); err != nil {
		return 0, nil, err
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBuffer, &jsonMap); err != nil {
		return 0, nil, err
	}

	return headerSize + 8, jsonMap, nil
}

func iterateEntries(jsonMap map[string]interface{}, callback func(map[string]interface{}, string) error) error {
	var helper func(map[string]interface{}, string) error
	helper = func(current map[string]interface{}, path string) error {
		if err := callback(current, path); err != nil {
			return err
		}
		if files, ok := current["files"].(map[string]interface{}); ok {
			for key, val := range files {
				if err := helper(val.(map[string]interface{}), filepath.Join(path, key)); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if files, ok := jsonMap["files"].(map[string]interface{}); ok {
		for key, val := range files {
			if err := helper(val.(map[string]interface{}), key); err != nil {
				return err
			}
		}
	}
	return nil
}


func extractFile(file *os.File, dst, path, offsetStr string, headerSize uint32, val map[string]interface{}) error {
	offset, err := strconv.ParseUint(offsetStr, 10, 64)
	if err != nil {
		return err
	}
	size := uint64(val["size"].(float64))
	if _, err := file.Seek(int64(headerSize)+int64(offset), io.SeekStart); err != nil {
		return err
	}
	buffer := make([]byte, size)
	if _, err := file.Read(buffer); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, path), buffer, 0644)
}

func Unpack(asar string, dst string) error {
	file, err := os.Open(asar)
	if err != nil {
		return err
	}
	defer file.Close()

	headerSize, jsonMap, err := readHeader(file)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return iterateEntries(jsonMap, func(val map[string]interface{}, path string) error {
		if offsetStr, ok := val["offset"].(string); ok {
			return extractFile(file, dst, path, offsetStr, headerSize, val)
		} else {
			return os.MkdirAll(filepath.Join(dst, path), 0755)
		}
	})
}

