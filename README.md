# hddgo
A library written in go to manipulate a harddrive.
I added also a wrapper to create luks or lvm container.
You cann create a lvm container in a luks Encrypted container or visa versa.



# Important Structs
0. Harddrive{}
0. Partition{}
0. LUKS{}
0. PhysicalVolume{}
0. VolumeGroup{}
0. LogicalVolume{}

# LVM
0. You can create Cachepool
1. You can move Physical Volumes 
2. You can create Snapshots
3. You can create Filesystem with parted

# Howto
0. Create a Partition: (h *Harddrive) CreateNewPartition(size string) (Partition, error)
0. Create a LUKS: (p *Partition) CreateLuksContainer(pwfile string) (LUKS, error)
0. Create a PhysicalVolume: (p *Partition) CreateNewPhysicalVolume()(error)
0. Create a VolumeGroup: (p *PhysicalVolume) CreateVolumeGroup(name string) (error)
0. Create a LogicalVolume: (v *VolumeGroup) CreateLogicalVolume(name string, size string) (error)


 
