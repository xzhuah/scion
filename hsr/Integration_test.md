# Prerequisites
Please refer to README.md for more information about HSR and DPDK.
Also, please refer to topology/mininet/README.md for instructions on setting up mininet.

# Running mininet
Here we assume a tiny topology.
```
[AS 13 servers]-[Mininet switch for AS13]-[tap]-[HSR in virtual box(AS 13 edge router)]-[tap]-[switch]-[AS 11 edge router]-[AS 11]-[AS 12]
```

## Install Mininet and dependencies
```
topology/mininet/setup.sh
```

## Run Mininet
```
./scion.sh topology zkclean -m -r -c topology/Tiny.topo
topology/mininet/run.sh -r
```

# Running HSR


## Install dependencies
This section explains the steps needed to setup a virtual machine environment. Once this is complete, you can start at the "Run VM" section for subsequent runs.

HSR requires Intel NICs supported by DPDK.
For testing we will use VirtualBox to emulate these NICs.

Before building a VM, the following software needs to be installed.
Default packages of Ubuntu 14.04 are a bit old, so you need to install them manually.
* Virtual Box version 5.0 (https://www.virtualbox.org/). Note that version 4.x does not support AES-NI, thus version 5.x is required.
* Vagrant (https://www.vagrantup.com/downloads.html) for building a VM with DPDK

Enable IP forwarding in the kernel
```
sudo sysctl net.ipv4.ip_forward=1
```

Download base ubuntu image.
```
vagrant box add --provider virtualbox bento/ubuntu-14.04
```

## Run VM

Create two taps. The VM uses eth10 and eth11 to communicate with mininet.
```
sudo ip tuntap add dev eth10 mode tap
sudo ip tuntap add dev eth11 mode tap
```
These devices will persist until the host is rebooted. To remove them manually, replace "add" with "del" in the above commands.

Build and start VM.
```
cd hsr/vagrant
vagrant up
```

Open a console.
```
vagrant ssh
```

At this point you are inside the VM. Frome there:

Build HSR
```
cd scion
./scion.sh build router
```

Setup DPDK environment.
```
cd hsr
./setup_dpdk.sh
```

## Run HSR
Inside the VM scion/hsr directory,

```
./exec.sh
```


## Test
Back in the host (separate terminal), run end2end_test
```
PYTHONPATH=. test/integration/end2end_test.py -m
```


<!--

# Modification of topology.py
In the mininet/topology.py eth10 and eth11 are connected with virtual switch s2 and s4, respectively (subject to change - refer to Troubleshooting section).
```
    for switch in net.switches:
        # switch.setMac("0:0:0:0:1:%x"%count)
        # count += 1
        if switch.name == "s2":
            Intf('eth10', node=switch)
        if switch.name == "s4":
            Intf('eth11', node=switch)

```

To disable the Python router (ER13), topology.py does not add link from/to er13.
```
    def addLink(self, node1, node2, params=None, intfName=None):
        if self.hsr:
            if node1 == "er1_13er1_11" or node2 == "er1_13er1_11":
                return
```

HSR does not support ARP, so hosts need to have static ARP entries.
topology.py executes arp command to insert the ARP entries. In the following case, HSR_EGRESS_IP and HSR_LOCAL_IP are IP addresses of HSR.
```
    for host in net.hosts:
        SNIP..
        if host.name == "er1_11er1_13":
            host.setMAC("0:0:0:0:0:CC", "er1_11er1_13-1")
        host.cmd("arp -s %s 1:2:3:4:5:6" % hsr_external_ip)
        host.cmd("arp -s %s 1:2:3:4:5:7" % hsr_internal_ip)

```


Moreover, topology.py executes following two commands.
```arp -s [hsr_internal_ip] 1:2:3:4:5:6``` (for sending ping packet to HSR)
```ip link set dev s4 addr 0:0:0:0:1:03``` (to fix the MAC address of switch s4. HSR uses this MAC address to send packet to end2end.py)
Note that mininet may change switch assignment, so please check which switch is for AS 13.
-->


# Troubleshooting
* Check that pox controller is installed in .local/bin.
```
$ which pox
/home/[your ID]/.local/bin/pox
```

* Start mininet before running the VM.
Sometimes packets from HSR are not delivered to Mininet hosts. In this case, please start Mininet first, then start the VM.

* Check MAC addresses of servers that are hardcoded in scion.c
```
 378   unsigned char mac_beacon[] = {0x0, 0x0, 0x0, 0x0, 0x0, 0xe};
 379   unsigned char mac_cert[] = {0x0, 0x0, 0x0, 0x0, 0x0, 0xd};
 380   unsigned char mac_path[] = {0x0, 0x0, 0x0, 0x0, 0x0, 0xf};
 381   unsigned char mac_dns[] = {0x0, 0x0, 0x0, 0x0, 0x0, 0x10};
```
You can get these MAC address from mininet console. For example,
```
mininet> bs1_13_1 ip link
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
63: bs1_13_1-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP mode DEFAULT group default qlen 1000
    link/ether 00:00:00:00:00:08 brd ff:ff:ff:ff:ff:ff
```


* Check static ARP entries of mininet hosts (HSR local/egress interfaces in topology.py)
* Check the MAC address of the AS 13 switch in topology.py.
```
os.system('ip link set dev s4 addr 0:0:0:0:1:03')  # HSR MAC address
```
Mininet sometimes changes switch assignment, so it may not be s4.

* Check that the emulated NICs(eth10 and eth11) are connected with mininet switch.
In the case below the switch of AS 13 is s1, thus eth11 is connected with s1.
Moreover, s4 is the switch between edge routers of AS11/AS13, thus eth10 is connected with s4.
```
*** Starting CLI:
mininet> net
bs1_11_1 bs1_11_1-0:s3-eth4
bs1_12_1 bs1_12_1-0:s0-eth1./de
bs1_13_1 bs1_13_1-0:s1-eth2
cs1_11_1 cs1_11_1-0:s3-eth2
cs1_12_1 cs1_12_1-0:s0-eth3
cs1_13_1 cs1_13_1-0:s1-eth1
ds1_11_1 ds1_11_1-0:s3-eth5
ds1_12_1 ds1_12_1-0:s0-eth4
ds1_13_1 ds1_13_1-0:s1-eth4
er1_11er1_12 er1_11er1_12-1:s2-eth2 er1_11er1_12-0:s3-eth1
er1_11er1_13 er1_11er1_13-0:s3-eth6 er1_11er1_13-1:s4-eth1
er1_12er1_11 er1_12er1_11-0:s0-eth2 er1_12er1_11-1:s2-eth1
er1_13er1_11
ps1_11_1 ps1_11_1-0:s3-eth3
ps1_12_1 ps1_12_1-0:s0-eth5
ps1_13_1 ps1_13_1-0:s1-eth3
s0 lo:  s0-eth1:bs1_12_1-0 s0-eth2:er1_12er1_11-0 s0-eth3:cs1_12_1-0 s0-eth4:ds1_12_1-0 s0-eth5:ps1_12_1-0
s1 lo:  s1-eth1:cs1_13_1-0 s1-eth2:bs1_13_1-0 s1-eth3:ps1_13_1-0 s1-eth4:ds1_13_1-0 eth11:
s2 lo:  s2-eth1:er1_12er1_11-1 s2-eth2:er1_11er1_12-1
s3 lo:  s3-eth1:er1_11er1_12-0 s3-eth2:cs1_11_1-0 s3-eth3:ps1_11_1-0 s3-eth4:bs1_11_1-0 s3-eth5:ds1_11_1-0 s3-eth6:er1_11er1_13-0
s4 lo:  s4-eth1:er1_11er1_13-1 eth10:
c0
```

In case that switch assignment is different, you may need to modify the following line in topology.py
```
        if switch.name == "s4":
            Intf('eth10', node=switch)
        if switch.name == "s1":
            Intf('eth11', node=switch)
```
