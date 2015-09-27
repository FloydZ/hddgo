package hdd

import (
	"os/exec"
)

func cryptsetupIsInstalled()(bool){
	_, err := exec.LookPath("cryptsetup")
	if err != nil {
		return false
	}
	return true
}
func partedIsInstalled()(bool){
	_, err := exec.LookPath("parted")
	if err != nil {
		return false
	}
	log.Info("Parted Found")
	return true
}


func isStringValidHDDorPart(s string) (bool, error){
	ret, err := isStringValidHarddrive(s)
	if err != nil {
		println("Error: " + err.Error())
		return false, err
	}
	if ret == true {return true, nil}


	ret, err = isStringValidPartition(s)
	if err != nil {
		println("Error: " + err.Error())
		return false, err
	}

	if ret == true {return true, nil}


	return false, nil
}

func isStringValidHarddrive(s string) (bool, error){
	array, err := InitAllHarddrives_Proc()
	if err != nil {
		println("Error: " + err.Error())
		return false, err
	}

	for _, a := range array{
		if a.Path == s {
			return true, nil
		}
	}

	return false, nil
}

func isStringValidPartition(s string) (bool, error){
	array, err := InitAllHarddrives_Proc()
	if err != nil {
		println("Error: " + err.Error())
		return false, err
	}

	for _, a := range array{
		if a.Path == s {
			return true, nil
		}
	}

	return false, nil
}

func isnumber(s rune) bool{
	switch {
	case s >= '0' && s <= '9':
		return true
	}

	return false
}

func clean(s string, what string) (string) {
	return s[len(what) + 2:len(s)-1]
}
