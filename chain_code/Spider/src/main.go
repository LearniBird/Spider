package main

import (
	"fmt"
	"os"
	"log"
	"io/ioutil"

	"net/http"
	"strings"
	"bytes"
	"strconv"
	"crypto/aes"
	"crypto/cipher"
	"time"
)


var(
	fileurl = ""
	passurl = ""

	url_suffix = ".ts"
	fileName = `.\src\video_src\course3\index.m3u8`
	fileLen = 0
	writeRootPath = `.\src\video_src\course3\`
	X_Requested_With = `ShockwaveFlash/31.0.0.108`
	User_Agent = `Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36`
	cookie = ``
)

func main() {
	contents := ReadFileM3U8(fileName)
	if contents == "" {
		os.Exit(1)
	}
	index := Index(contents)
	fileLen = len(index[0])
	fmt.Println("download begin.....")

	pass, _ := httpResp(passurl)
	fmt.Printf("pass %s\n",pass)
	time.Sleep(10)
	i := 0
	content := make([][]byte, fileLen)
	for i, _ = range index[0] {
		fmt.Println(fmt.Sprintf(fileurl,index[0][i], index[1][i]))
		content[i] = Download(fmt.Sprintf(fileurl,index[0][i], index[1][i]), i, pass)
	}
	fmt.Printf("len(index): %d\n", i)
	fmt.Println("download finish.")
	sep := []byte("")
	mergeBytes := bytes.Join(content,sep)
	f, err := os.Create(writeRootPath + "merge.ts") //创建文件
	if err != nil {
		fmt.Println("create %s failed!",writeRootPath + "merge.avi" )
		os.Exit(1)
	}
	defer f.Close()
	_, err = f.Write(mergeBytes)
	if err != nil {
		fmt.Printf("Wtite %s failed! \n", writeRootPath + "merge.avi")
		os.Exit(1)
	}
	fmt.Printf("Wtite %s success! \n", writeRootPath + "merge.avi")
}

func aesDecrypty(result []byte, key []byte) []byte {
	origData, err := AesDecrypt(result, key)
	if err != nil {
		panic(err)
	}
	return origData
}

func httpResp(url string) ([]byte, int) {
	client := &http.Client{}
	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)

	//增加header选项
	//reqest.Header.Add("Accept-Encoding", "gzip, deflate, br")
	reqest.Header.Add("Accept", "*/*")
	//reqest.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	reqest.Header.Add("Cookie", cookie)
	reqest.Header.Add("Connection", "keep-alive")
	reqest.Header.Add("Host", "vod2.xiaoe-tech.com")
	reqest.Header.Add("Referer", "https://pc-shop.xiaoe-tech.com/appBbtPFWyR9948/video_details?id=v_5ba5afd35ac1e_d98Q8HTG")
	reqest.Header.Add("User-Agent", User_Agent)
	reqest.Header.Add("X-Requested-With", X_Requested_With)

	if err != nil {
		panic(err)
	}
	//处理返回结果
	response, _ := client.Do(reqest)
	body, err := ioutil.ReadAll(response.Body)
	return body, response.StatusCode

}


func Download(url string,i int, key []byte) []byte {
	content, statusCode := httpResp(url)
	if statusCode != 200 {
		fmt.Printf("status code : %d \n", statusCode)
		os.Exit(1)
	}
	content = aesDecrypty(content, key)
	strInx := strconv.Itoa(i)
	fullName := writeRootPath + strInx + ".ts"
	f, err := os.Create(fullName) //创建文件
	if err != nil {
		fmt.Printf("create %s failed! \n", fullName)
		os.Exit(1)
	}
	defer f.Close()
	_, err = f.Write(content)
	if err != nil {
		fmt.Printf("Wtite %s failed! \n", fullName)
		os.Exit(1)
	}
	fmt.Printf("Wtite %s success! \n", fullName)
	//fmt.Printf("content: %s \n\n",content)
	return content
}

func Index(contents string) [2][]string {
	var index  [2][]string

	for {
		i :=strings.Index(contents, "start=")
		if i== -1 {
			break
		}
		contents = contents[i+6:]
		j := 0
		for {
			if contents[j] == '&' {
				break
			}
			j++
		}
		index[0] = append(index[0], contents[0:j] )
		contents = contents[j+5:]
		j = 0
		for {
			if contents[j] == '&' {
				break
			}
			j++
		}
		index[1] = append(index[1], contents[0:j] )
	}


	return index
}

func ReadFileM3U8(fileName string) string {
	if FileExists(fileName) {
		contents, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Panic(err)
		}
		return string(contents[:])
	}
	return ""
}

func FileExists(fileName string) bool {
	if _,err :=os.Stat(fileName); os.IsNotExist(err){
		log.Panic(err)
		os.Exit(1)
	}
	return true
}

func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}


func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))

	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)

	return origData, nil
}

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
