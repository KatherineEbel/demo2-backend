package demo2fs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func CreateFile(path string, name string) (string, error) {
	var _, err = os.Stat(path + name)
	if os.IsNotExist(err) {
		file, err := os.Create(path + name)
		if err != nil {
			defer func() {
				_ = file.Close()
			}()
			return "", err
		}
	}
	return path + name, nil
}

func ReadFile(path string) ([]byte, error) {
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return nil, err
	}
	var data = make([]byte, 1024)
	for {
		_, err = file.Read(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
	}
	fmt.Println("==> done reading from file")
	return data, nil
}

func ReadFile2(fp string) ([]byte, error) {
	file, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	err = nil
	buf := make([]byte, 32*1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			fmt.Print(buf[:n]) // this prints the data out when done...
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("read %d bytes: %v", n, err)
			break
		}
	}
	buf = bytes.Trim(buf, "\x00")
	return buf, err
}
func WriteFile(path string, data []byte) error {
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		defer func() {
			_ = file.Close()
		}()
		return err
	}
	_ = file.Truncate(0)
	_, _ = file.Seek(0, 0)
	if err = ioutil.WriteFile(file.Name(), data, 0664); err != nil {
		return err
	}
	fmt.Println("===> done writing to file")
	return nil
}

func DeleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		return err
	}
	fmt.Println("===> done deleting file")
	return nil
}
