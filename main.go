package main

import (
    "fix-sources-list/source"
    "fix-sources-list/utils"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "os"
    "strings"
    "time"
)

//noinspection GoSnakeCaseUsage
const (
    SOURCES_LIST_PATH = "/etc/apt/sources.list"
)

type IntArray []int

func (arr IntArray) Contains(i int) bool {
    for _, j := range arr {
        if j == i {
            return true
        }
    }
    return false
}

func main() {
    log.Println("Reading file...")
    fInfo, err := os.Lstat(SOURCES_LIST_PATH)
    utils.CheckErr(err)
    content, err := ioutil.ReadFile(SOURCES_LIST_PATH)
    utils.CheckErr(err)
    log.Println("Checking writing permissions...")
    utils.CheckErr(ioutil.WriteFile(SOURCES_LIST_PATH, content, fInfo.Mode()))
    
    log.Println("Patching sources...")
    entries := make(map[int]source.Entry)
    mergedLines := make(IntArray, 0)
    lines := strings.Split(string(content), "\n")
    for ln, line := range lines {
        line = strings.TrimSpace(line)
        if len(line) != 0 && []rune(line)[0] != '#' {
            merged := false
            
            for mLn, entry := range entries {
                if !merged && entry.CanMergeWith(source.Entry(line)) {
                    entries[mLn] = entry.MergeComponents(source.Entry(line).Components()...)
                    lines[ln] = fmt.Sprintf("#-MERGED-(%d)-#%s", mLn, line)
                    merged = true
                    if !mergedLines.Contains(mLn) {
                        lines[mLn] = "#-MERGED-#" + lines[mLn]
                        mergedLines = append(mergedLines, mLn)
                    }
                }
            }
            
            if !merged {
                entries[ln] = source.Entry(line)
            }
        }
    }
    
    now := time.Now()
    if len(mergedLines) > 0 {
        lines = append(lines, "", "", fmt.Sprintf(
            "############################ Merged Lines (%s) ############################", now.Format(time.RFC1123)))
        
        for _, ln := range mergedLines {
            lines = append(lines, string(entries[ln]))
        }
        
        log.Println("Creating backup...")
        CreateBackup(fInfo.Mode())
        
        log.Println("Replacing old file...")
        utils.CheckErr(ioutil.WriteFile(SOURCES_LIST_PATH, []byte(strings.Join(lines, "\n")), fInfo.Mode()))
    }
    log.Println("Lines fixed: ", len(mergedLines))
}

func CreateBackup(perm os.FileMode) {
    buFile, err := os.OpenFile(SOURCES_LIST_PATH + ".bak", os.O_CREATE | os.O_TRUNC | os.O_WRONLY, perm)
    utils.CheckErr(err)
    defer buFile.Close()
    file, err := os.Open(SOURCES_LIST_PATH)
    utils.CheckErr(err)
    defer file.Close()
    _, err = io.Copy(buFile, file)
    utils.CheckErr(err)
}