package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
)

const (
	output_file = "master_list.txt"
	final_file  = "final.csv"
)

func Determineos() string {
	switch runtime.GOOS {
	case "darwin", "dragonfly", "freebsd", "linux", "netbsd", "openbsd", "plan9", "solaris":
		return "nix"
	case "windows":
		return "windows"
	}
	return "unknown os: " + runtime.GOOS
}

//Readline
func Readline(file string) []string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal("File error %s", err)
	}
	return strings.Split(string(content), "\n")
}

//Save
func Save(filename string, lines []string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("Save output error %s", err)
	}
	defer f.Close()

	for _, v := range lines {
		fmt.Fprintf(f, "%s\n", v)
	}
}

//Uniqe lines
func Uniq(lines []string) []string {
	var ulines []string
	set := make(map[string]bool)
	for _, v := range lines {
		if !set[v] {
			ulines = append(ulines, v)
			set[v] = true
		}
	}
	return ulines
}

//Readline of file1 and file2 and write uniq into master_list.txt
func Filecompare(file1, file2 string) []string {
	var full_final []string

	f1 := Readline(file1)
	full_final = append(full_final, f1...)

	f2 := Readline(file2)
	full_final = append(full_final, f2...)

	//Uniq
	full_final = Uniq(full_final)

	//Save
	Save(output_file, full_final)

	return full_final
}

func main() {
	var full_infos []string

	flag.Parse()
	if len(flag.Args()) < 2 {
		log.Fatal("Usage:\n  tactilefileprofile <file1> <file2>")
	}

	os_decided := Determineos()
	file_manip := Filecompare(flag.Arg(0), flag.Arg(1))

	//Only nix
	if os_decided == "nix" {
		for _, pelement := range file_manip {
			var linfo, pfsgid, pfsuid, pwwd string
			fileInfo, err := os.Stat(pelement)
			if err == nil {
				stat := fileInfo.Sys().(*syscall.Stat_t)

				//setuid
				if stat.Mode&syscall.S_ISUID != 0 {
					pfsuid = "Y"
				} else {
					pfsuid = "N"
				}

				//setuid
				if stat.Mode&syscall.S_ISGID != 0 {
					pfsgid = "Y"
				} else {
					pfsgid = "N"
				}

				//Sticky
				if stat.Mode&syscall.S_ISVTX != 0 {
					pwwd = "Y"
				} else {
					pwwd = "N"
				}

				pacl := fileInfo.Mode().Perm()
				pfuser := stat.Uid
				pfgroup := stat.Gid

				if !fileInfo.IsDir() {
					//File
					linfo = fmt.Sprintf("Y|%s|File||Y|%o|%d|%d|%s|%s|NULL|NULL|NULL|NULL|NULL|NULL|NULL|NULL|NULL|NULL", pelement, pacl, pfuser, pfgroup, pfsuid, pfsgid)
				} else {
					//Directory
					linfo = fmt.Sprintf("Y|%s|Directory||N|NULL|NULL|NULL|%s|%s|%o|%d|%d|%s|NULL|NULL|NULL|NULL|NULL|NULL", pelement, pfsuid, pfsgid, pacl, pfuser, pfgroup, pwwd)
				}
				full_infos = append(full_infos, linfo)
			}

		}
	}

	Save(final_file, full_infos)

}
