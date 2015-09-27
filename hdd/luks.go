package hdd
import (
	"os/exec"
	"strings"
	"errors"
)


type LUKS struct {
	Part 					Partition		// Für .Directory
	PhysicalVolumeName 		string	// Für den Namen
	Hash 					string
	Cipher 					string
	KeyFile 				string
	UsedKeySlot 			int
	Suspended				bool
}

func (l *LUKS) LuksFormat() (error){
	if l.Part.Path == "" {return errors.New("Please set a Partition")}

	keycmd := ""
	if l.KeyFile != "" { keycmd = " " + l.KeyFile}

	if l.Cipher == "" {l.Cipher = "aes-xts-plain64"}

	cmd := sudo + " cryptsetup " + " -c " + l.Cipher + " -y " + "-s " + "512 " + "luksFormat " + l.Part.Path + keycmd
	log.Info(cmd)


	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	ret, err := exec.Command(head, parts...).Output()
	if err != nil {
		println("Error: " + err.Error())
		println(string(ret))
		return err
	}
	return nil
}
func (l *LUKS) LuksOpen() (error){
	if l.PhysicalVolumeName == "" {return errors.New("Please set a PV Name")}
	if l.Part.Path == "" {return errors.New("Please set a Partition")}

	keycmd := ""
	if l.KeyFile != "" { keycmd = " --key-file " + l.KeyFile}

	cmd := sudo + " cryptsetup luksOpen " + l.Part.Path + " " + l.PhysicalVolumeName + " " + keycmd
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	ret, err := exec.Command(head, parts...).Output()
	if err != nil {
		println("Error: " + err.Error())
		println(string(ret))
		return err
	}
	return nil
}
func (l *LUKS) LuksSuspend() (error){
	if l.PhysicalVolumeName == "" {return errors.New("Please set a PV Name")}

	keycmd := ""
	if l.KeyFile != "" { keycmd = " --key-file " + l.KeyFile}

	cmd := sudo + " cryptsetup luksSuspend " + l.Part.Path + keycmd
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	ret, err := exec.Command(head, parts...).Output()
	if err != nil {
		println("Error: " + err.Error())
		println(string(ret))
		return err
	}

	l.Suspended = true
	return nil
}
func (l *LUKS) LuksResume() (error){
	if l.PhysicalVolumeName == "" {return errors.New("Please set a PV Name")}
	if l.Suspended == false{return errors.New("Is already up")}

	keycmd := ""
	if l.KeyFile != "" {keycmd = " " + l.KeyFile}

	cmd := sudo + " cryptsetup luksResume " + l.PhysicalVolumeName  + keycmd
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	ret, err := exec.Command(head, parts...).Output()
	if err != nil {
		println("Error: " + err.Error())
		println(string(ret))

		return err
	}

	l.Suspended = false
	return nil
}
func (l *LUKS) LuksErase() (error){
	if l.Part.Path == "" {return errors.New("Please set a Partition")}

	cmd := sudo + " cryptsetup luksErase " + l.Part.Path
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	ret, err := exec.Command(head, parts...).Output()
	if err != nil {
		println("Error: " + err.Error())
		println(string(ret))

		return err
	}
	return nil

}
func (l *LUKS) LuksChangeKey(newKeyFile string)(error){
	if l.Part.Path == "" {return errors.New("Please set a Partition")}
	if newKeyFile == "" {return errors.New("Please set a Partition")}

	keycmd := ""
	if l.KeyFile != "" { keycmd = " --key-file " +l.KeyFile}

	cmd := sudo + " cryptsetup luksChangeKey " + l.Part.Path + keycmd + " " + newKeyFile
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	ret, err := exec.Command(head, parts...).Output()
	if err != nil {
		println("Error: " + err.Error())
		println(string(ret))

		return err
	}

	l.KeyFile = newKeyFile
	return nil
}
func (l *LUKS) LuksAddKey(newKeyFile string)(error){
	if l.PhysicalVolumeName == "" {return errors.New("Please set a PV Name")}
	if newKeyFile == "" {return errors.New("Please set a new Keyfile")}

	keycmd := ""
	if l.KeyFile != "" { keycmd =  " --key-file " + l.KeyFile}

	cmd := sudo + " cryptsetup luksAddKey " + l.Part.Path + keycmd + " " + newKeyFile
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	ret, err := exec.Command(head, parts...).Output()
	if err != nil {
		println("Error: " + err.Error())
		println(string(ret))
		return err
	}
	return nil
}
func (l *LUKS) LuksRemoveKey(keyfile string)(error){
	if l.PhysicalVolumeName == "" {return errors.New("Please set a PV Name")}
	if keyfile == "" {return errors.New("Please set a keyfile to delete")}

	keycmd := ""
	if l.KeyFile != "" { keycmd =  " --key-file " + l.KeyFile}

	cmd := sudo + " cryptsetup luksRemoveKey " + l.Part.Path + keycmd + " " + keyfile
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		println("Error: " + err.Error())
		return err
	}
	return nil
}
func luksKillSLot(){}
func luksUUID(){}
func luksDumo(){}
func isLuks(){}
