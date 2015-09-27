package hdd

import (
	"github.com/op/go-logging"
	linuxproc "github.com/c9s/goprocinfo/linux"							//zum Auslesen von /proc/...
	"strings"
	"errors"
	"strconv"
	"unicode"
	"github.com/pivotal-golang/bytefmt"
	"unicode/utf8"
	"os/exec"
	"fmt"
)

var sudo = "sudo"
var lsblkcmd = "+UUID,OWNER,GROUP,MODE,STATE,HOTPLUG,FSTYPE,LABEL,MODEL,SERIAL"

var log = logging.MustGetLogger("")

type Harddrive struct {
	Id 				int
	Name 			string			// bsp: sda1
	Path 			string 			// bsp: /dev/sda1
	Diskstat 		linuxproc.DiskStat

	Size 			string 			//379,7G
	SizeRead 		string 			//823123189237

	Free 			string			//379G
	FreeRead 		string 			//8189237

	Part 			[]Partition

	Vendor			string			// Hersteller
	Model 			string			//
	Serial 			string
	Status 			string 			//e.g. running
	Label 			string
	FSType 			string			//e.g. ext4, LVM2_member, swap
	UUID			string			//e.g. 8a7c2183-f8e4-4c20-8f5a-d7cb1c6196b6
	Owner 			string
	Group 			string
	Mode 			string			//e.g. brw-rw----
	Hotplug 		bool 			//e.g. plug n play

	//Für Parted
	LastByte		string 			//e.g. 763832312
	LastByteRead	string 			//e.g. 379G //For new Partition

}



func (h *Harddrive) CreateNewPhysicalVolume(){
	p, err := CreateNewPhysicalVolumeFromString(h.Path, "")
	if err != nil {
		log.Error(err.Error())
		return
	}

	//%TODO add to an array
	fmt.Println(p)

	return

}

func (h *Harddrive) CreateNewPhysicalVolumeWithCMD(cmd string){
	p, err := CreateNewPhysicalVolumeFromString(h.Path, cmd)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//%TODO add to an array
	fmt.Println(p)

	return

}

func (h *Harddrive) SetLabel(label string, index int) (error){
	cmd := sudo + " parted " + h.Path + " name " + strconv.Itoa(index) + " " + label
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return  err
	}

	return nil
}

func (h *Harddrive) GetLastByte() (error){
	cmd := sudo + " parted " + h.Path + " print"
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	ret, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error() + " Return:" + string(ret))
		return err
	}

	lines := strings.Split(string(ret), "\n")
	fields := strings.Fields(lines[len(lines) - 3])
	h.LastByteRead = fields[2][:len(fields[2])- 1]
	h.LastByte = convertSizeReverse(fields[2][:len(fields[2])- 1])

	log.Info("LastByte: " + h.LastByte)
	//println("getlastbyte:" + h.LastByteRead)
	//println("getlastbyte:" + h.LastByte)
	//println("getlastbyte:" + convertSize(h.LastByte))
	//%TODO Check for right format
	return nil
}

func (h *Harddrive) GetFreeSpace() (){
	//%TODO
	h.GetLastByte()
	if h.HasFreeSpace() == false {return}
	println(subtractSizeStrings(h.Size, h.LastByte))
	return
}

func (h *Harddrive) HasFreeSpace() (bool){
	if h.LastByte == "" {h.GetLastByte()}
	if h.LastByte == h.Size {return false}
	return true
}

func (h *Harddrive) CreateNewPrimaryPartitionOnPostion(start string, end string) (Partition, error){
	p := Partition{}
	if parseSize(start) == false {return p, errors.New("Please supply a valid size format")}
	if parseSize(end) == false {return p, errors.New("Please supply a valid size format")}

	log.Info("Start:" + start)
	log.Info("End:" + end)


	//%TODO WICHTIG!!!!!! B nicht vergessen, wird in der struct fallen gelasen
	cmd := sudo + " parted " + h.Path + " mkpart primary " + start + " " + end
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	//Debug
	return p, nil
	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return p, err
	}

	return p, nil
}

func (h *Harddrive) CreateNewPartition(size string) (Partition, error){
	p := Partition{}
	if parseSize(size) == false {return p, errors.New("Please supply a valid size format")}

	log.Info("Size:" + size)

	p, err := h.CreateNewPrimaryPartitionOnPostion(h.LastByteRead , addSizeStrings_Read(h.LastByteRead, size))
	if err != nil {
		log.Error(err.Error())
		return p, err
	}

	return p, nil
}

func (h *Harddrive) GetAllPartitions() (error){
	var x = 0
	funcf := func(c rune) bool {
		if (c == '"'){
			x = x + 1
			x = x % 2
		}
		if (x == 0){ return unicode.IsSpace(c)}	 // " is nicht offen
		return false							 // ein " ist offen
	}

	Array := []Partition{}
	data, err := exec.Command("lsblk", "-P", "-o", lsblkcmd, h.Path).Output()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	out := string(data)
	lines := strings.Split(out, "\n")

	for _, line := range lines {
		if line == "" {break;}
		fields := strings.FieldsFunc(line, funcf)

		if !(clean(fields[0],"NAME") == "NAME"){
			if (clean(fields[5],"TYPE") == "part") {//Partition
				e := fields[0][len(fields[0])-2:len(fields[0])-1]

				//println("e:" + string(e))
				//println("field9:" + fields[0])

				a := linuxproc.DiskStat{Name:clean(fields[0], "NAME")}
				p := Partition{Number:e, Name:clean(fields[0], "NAME"), Path:"/dev/" + clean(fields[0], "NAME"), Diskstat:a,
					MountPoint:clean(fields[6], "MOUNTPOINT"), UUID:clean(fields[7], "UUID"), Owner:clean(fields[8], "OWNER"), Group:clean(fields[9], "GROUP"),
					Mode:clean(fields[10], "MODE"), Type:clean(fields[5], "TYPE"), Size:clean(fields[3], "SIZE"), Status:clean(fields[11], "STATE"),
					FSType:clean(fields[13], "FSTYPE"), Label:clean(fields[14], "LABEL")}

				Array = append(Array, p)
			}
		}
	}

	//fmt.Println(Array)
	h.Part = Array
	return nil
}

func (h *Harddrive) init() (){
	//So ne funktion um alles felder der structur zu setzen nur intern
	//h.GetLastByte()
	//%TODO
	return
}





func GetAllHarddrivesAsStrings() ([]string, error){
	Array, err := GetAllHarddrives()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	ret := []string{}
	for _, a := range Array {
		ret = append(ret, a.Path)
	}

	return ret, nil
}

func GetAllHarddrives() ([]Harddrive, error){
	var x = 0
	funcf := func(c rune) bool {
		if (c == '"'){
			x = x + 1
			x = x % 2
		}
		if (x == 0){ return unicode.IsSpace(c)}	 // " is nicht offen
		return false							 // ein " ist offen
	}

	Array := []Harddrive{}
	data, err := exec.Command("lsblk", "-dPb", "-o", lsblkcmd).Output()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	out := string(data)
	lines := strings.Split(out, "\n")

	for _, line := range lines {
		if line == "" {break;}
		fields := strings.FieldsFunc(line, funcf)

		if !(clean(fields[0],"NAME") == "NAME"){
			hotplug := false
			if clean(fields[12],"HOTPLUG") == "0"{hotplug = false} else {hotplug = true}

			a := linuxproc.DiskStat{Name:clean(fields[0],"NAME")}//%TODO

			h := Harddrive{Name:clean(fields[0],"NAME"), Model:clean(fields[15],"MODEL"), Serial:clean(fields[16],"SERIAL"), Path:"/dev/" + clean(fields[0],"NAME"), Diskstat:a,
				UUID:clean(fields[7],"UUID"), Owner:clean(fields[8],"OWNER"), Group:clean(fields[9],"GROUP"), Hotplug:hotplug,
				Mode:clean(fields[10],"MODE"), Size:convertSize(clean(fields[3],"SIZE")), SizeRead:clean(fields[3],"SIZE"), Status:clean(fields[11],"STATE")}

			h.GetAllPartitions()
			h.init()
			Array = append(Array, h)
		}
	}

	//fmt.Println(Array[0].Size)
	return Array, nil
}

func GetHarddriveFromString(Harddisk string) (Harddrive, error){
	h :=Harddrive{}

	array, err := GetAllHarddrives()
	if err != nil {
		log.Error(err.Error())
		return h, err
	}

	for _, a := range array {
		if a.Path == Harddisk{
			h = a
		}
	}
	if h.Name == "" {
		log.Error("No Harddrive with Name: " +  Harddisk + "found")
		return h, errors.New("No Harddrive with Name: " +  Harddisk + "found")
	}
	return h, nil
}

func InitAllHarddrives_Proc() ([]Harddrive, error){

	DriveArray := []Harddrive{}

	a, err := linuxproc.ReadDiskStats("/proc/diskstats")
	if err != nil{
		log.Error(err.Error())
		return nil, err
	}

	for i := 0;i < len(a) ;i++  {
		end := a[i].Name[len(a[i].Name)-1:]
		b, _ := utf8.DecodeRuneInString(end)
		if isnumber(b) == false{//HDD
			h := Harddrive{Path:"/dev/"+a[i].Name, Name:a[i].Name, Diskstat:a[i]}
			PartArray := []Partition{}
			counter2 := 1;
			for {
				if (i + counter2) >= len(a){break;}
				e := a[i + counter2].Name[len(a[i + counter2].Name)-1:]
				b, _ := utf8.DecodeRuneInString(e)
				if isnumber(b) == false {//HDD
					h.Part = PartArray
					break
				}
				if isnumber(b) == true {//Partition
					p := Partition{Number:e, Name:a[i + counter2].Name, Path:"/dev/"+a[i + counter2].Name, Diskstat:a[i + counter2]}
					PartArray = append(PartArray, p)
				}
				counter2 += 1
			}
			DriveArray = append(DriveArray, h)
		}

	}
	return DriveArray,nil

}




func parseSize(s string)(bool){
	//%TODO
	return true
}

func convertSizeReverse(s string)(string){
	// aus 962.1G wird 314572800000
	s = strings.Replace(s, ",", ".", -1)
	pos := strings.Index(s, ".")

	part1 := " "
	part2 := "1"

	if pos == -1{
		part1 = s[:len(s)-1]
	}else{
		part1 = s[:pos]
		part2 = s[pos+1:pos+2]

	}
	part3 := s[len(s)-1:]

	if len(part1) < 1 { return ""}
	//println("Part1:" + part1)
	//println("Part2:" + part2)
	//println("Part3:" + part3)

	value, err := strconv.ParseUint(part1, 10, 0)
	if err != nil || value < 1 {
		log.Error(err.Error())
		return ""
	}
	value2, err := strconv.ParseUint(part2, 10, 0)
	if err != nil || value < 1 {
		log.Error(err.Error())
		return ""
	}

	var bytes uint64
	var bytes2 uint64

	unit := strings.ToUpper(part3)
	switch unit[:1] {
	case "T":
		bytes = value * bytefmt.TERABYTE
		bytes2= value2* bytefmt.GIGABYTE
	case "G":
		bytes = value * bytefmt.GIGABYTE
		bytes2= value2* bytefmt.MEGABYTE
	case "M":
		bytes = value * bytefmt.MEGABYTE
		bytes2= value2* bytefmt.KILOBYTE
	case "K":
		bytes = value * bytefmt.KILOBYTE
		bytes2= value2* bytefmt.BYTE
	case "B":
		bytes = value * bytefmt.BYTE
		bytes2= value2
	default:
		bytes = 0
	}

	ret := strconv.FormatUint(bytes + bytes2, 10)
	log.Info("Aus " + s + " wurde " + ret)
	return ret
}

func convertSize(s string)(string){
	// aus 314572800000 wird 962.1G
	ret, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Error(err.Error())
		return ""
	}
	str := bytefmt.ByteSize(ret)
	log.Info("Aus " + s + " wurde " + str)
	return str
}

func subtractSizeStrings(a, b string)(string){
	//Input 123123123 und 2323

	sizeI, err := strconv.ParseUint(a, 10, 0)
	if err != nil || sizeI < 1 {
		log.Error(err.Error())
		return ""
	}

	usedI, err := strconv.ParseUint(b, 10, 0)
	if err != nil || usedI < 1 {
		log.Error(err.Error())
		return ""
	}

	freeI := strconv.FormatUint(sizeI - usedI, 10)
	log.Info("Aus " + a + " - " + b + " wurde " + freeI)
	return freeI
}

func subtractSizeStrings_Read(a, b string)(string){
	//Input 345,3G und 234G
	size := convertSizeReverse(a)
	used := convertSizeReverse(b)

	//println("size:" + size + " used" + used )

	sizeI, err := strconv.ParseUint(size, 10, 64)
	if err != nil || sizeI < 1 {
		log.Error(err.Error())
		return ""
	}

	usedI, err := strconv.ParseUint(used, 10, 64)
	if err != nil || usedI < 1 {
		log.Error(err.Error())
		return ""
	}

	freeI := convertSize(strconv.FormatUint(sizeI - usedI, 10))
	log.Info("Aus " + a + " - " + b + " wurde " + freeI)

	return freeI
}

func addSizeStrings_Read(a, b string)(string){
	//Input 345,3G und 234G
	//println("addSizeStrings_Read: size:" + a + " used:" + b )

	size := convertSizeReverse(a)
	used := convertSizeReverse(b)

	if (size == "") || (used == ""){return ""}
	ret := convertSize(addSizeStrings(size, used))
	log.Info("Aus " + a + " + " + b + " wurde " + ret)
	return ret
}

func addSizeStrings(a, b string)(string){
	//Input 3234872348 und 12355

	sizeI, err := strconv.ParseUint(a, 10, 64)
	if err != nil || sizeI < 1 {
		log.Error(err.Error())
		return ""
	}

	usedI, err := strconv.ParseUint(b, 10, 64)
	if err != nil || usedI < 1 {
		log.Error(err.Error())
		return ""
	}

	freeI := strconv.FormatUint(sizeI + usedI, 10)
	log.Info("Aus " + a + " + " + b + " wurde " + freeI)

	return freeI
}


