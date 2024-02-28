package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	filelList, err := getFileList("./input")
	var bnkFilelList []string
	var wemFilelList []string
	for _, file := range filelList {
		if file[len(file)-3:] == "bnk" {
			bnkFilelList = append(bnkFilelList, file)
		} else if file[len(file)-3:] == "wem" {
			wemFilelList = append(wemFilelList, file)
		}
	}
	if err != nil {
		fmt.Println("Can not get bnk list", err)
	}
	var failed []string
	count := 0
	taskNumber := len(filelList)
	for _, bnkFile := range bnkFilelList {
		fmt.Printf("%d/%d    ", count, taskNumber)
		count += 1
		err = bnkProcess(bnkFile)
		if err != nil {
			failed = append(failed, bnkFile)
		}
	}
	for _, wemFile := range wemFilelList {
		fmt.Printf("%d/%d", count, taskNumber)
		count += 1
		err = wemProcess(wemFile)
		if err != nil {
			failed = append(failed, wemFile)
		}
	}
	fmt.Println("失败列表", failed)
	fmt.Scanln()
}

// bnk处理
func bnkProcess(bnkFile string) error {
	fmt.Println("Processing", bnkFile)

	if err := initWorkspace(); err != nil {
		fmt.Println("initWorkspace 出错:", err)
		return err
	}

	if err := bnk2wav(bnkFile); err != nil {
		fmt.Println("bnk2wav 出错:", err)
		return err
	}

	if err := wav2wem(bnkFile[6:]); err != nil {
		fmt.Println("wav2wem 出错:", err)
		return err
	}

	if err := wem2ogg(); err != nil {
		fmt.Println("wem2ogg 出错:", err)
		return err
	}

	if err := mvOgg(); err != nil {
		fmt.Println("mvOgg 出错:", err)
		return err
	}
	fmt.Println("处理完成")
	return nil
}

// wem处理
func wemProcess(wemFile string) error {
	fmt.Println("Processing", wemFile)

	if err := initWorkspace(); err != nil {
		fmt.Println("initWorkspace 出错:", err)
		return err
	}
	if err := copyFile(wemFile, "./work"+wemFile[5:]); err != nil {
		fmt.Println("wem文件复制错误")
	}
	if err := wem2ogg(); err != nil {
		fmt.Println("wem2ogg 出错:", err)
		return err
	}

	if err := mvOgg(); err != nil {
		fmt.Println("mvOgg 出错:", err)
		return err
	}
	fmt.Println("处理完成")
	return nil
}

// 初始化工作目录
func initWorkspace() error {
	err := os.RemoveAll("./work")
	if err != nil {
		fmt.Println("Can not clean up work diectory:", err)
		return err
	}
	err = os.Mkdir("./work", 0755)
	if err != nil {
		fmt.Println("Can not create work diectory:", err)
	}
	return nil
}

// 复制文件
func copyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destinationFile.Close()
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}

// 获取Bnk文件列表
func getFileList(dirPath string) ([]string, error) {
	var fileList []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		fileList = append(fileList, path)
		return nil
	})
	return fileList, err
}

// 把Bnk文件转为加密的wav文件
func bnk2wav(bnkFile string) error {
	copyFile("./bnkextr.exe", "./work/bnkextr.exe")
	defer os.Remove("./work/bnkextr.exe")
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Cannot get current dir", err)
	}
	bnkextr := "./bnkextr.exe"
	cmd := exec.Command(bnkextr, filepath.Join(currentDir, bnkFile))
	cmd.Dir = "./work"
	err = cmd.Start()
	if err != nil {
		fmt.Println("Bnk extract start failed:", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println("Bnk extract process error:", err)
		return err
	}
	return nil
}

// wav文件重命名为wem文件
func wav2wem(name string) error {
	wavList, err := getFileList("./work")
	if err != nil {
		fmt.Println("Can not get wav file list:", err)
	}

	for _, wav := range wavList {
		// fmt.Println(wav)
		err := os.Rename(wav, filepath.Join("./work", name[:len(name)-4]+"--"+wav[6:len(wav)-4]+".wem"))
		if err != nil {
			fmt.Println("Can not rename wav file")
		}
	}
	return nil
}

// wem文件转为ogg文件
func wem2ogg() error {
	wemList, err := getFileList("./work")
	if err != nil {
		fmt.Println("Can not get wem file List:", err)
		return err
	}
	err = copyFile("./ww2ogg.exe", "./work/ww2ogg.exe")
	if err != nil {
		fmt.Println("Can not copy ww2ogg.exe", err)
		return err
	}
	err = copyFile("./packed_codebooks_aoTuV_603.bin", "./work/packed_codebooks_aoTuV_603.bin")
	if err != nil {
		fmt.Println("Can not copy packed_codebooks_aoTuV_603.bin", err)
		return err
	}
	defer os.Remove("./work/ww2ogg.exe")
	defer os.ReadDir("./work/packed_codebooks_aoTuV_603.bin")
	for _, wem := range wemList {
		// fmt.Println(wem[5:])
		cmd := exec.Command("./ww2ogg.exe", "--pcb", "./packed_codebooks_aoTuV_603.bin", "./"+wem[5:])
		cmd.Dir = "./work"
		err = cmd.Start()
		if err != nil {
			fmt.Println("wem to ogg failed:", err)
			return err
		}
		err = cmd.Wait()
		if err != nil {
			fmt.Println("wem to ogg process error:", err)
			return err
		}
	}
	return nil
}

// ogg文件放到输出目录
func mvOgg() error {
	fileList, err := getFileList("./work")
	if err != nil {
		fmt.Println("Can get file list of work directory:", err)
		return err
	}
	var oggList []string
	for _, ogg := range fileList {
		// fmt.Println(ogg[len(ogg)-3:])
		if ogg[len(ogg)-3:] == "ogg" {
			oggList = append(oggList, ogg)
		}
	}
	for _, ogg := range oggList {
		err = os.Rename(ogg, "./output/"+ogg[5:])
		if err != nil {
			fmt.Println("Can not move ogg file:", err)
			return err
		}
	}

	return nil
}
