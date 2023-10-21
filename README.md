# gopatate

# A word 

This tool is 
* An simple brute force tool and will be stay this way
* Useful for basic stuff
* Free of any other library to avoid problem with compliance shit

Yes you can do it yourself. 
Yes you may do better.
Yes you can improve it in many ways, BUT remember to avoid dependencies of non-native library.

# Install

```bash
go get install github.com/po0lpy0x0c/gopatate@latest

# You add your go/bin folder to your binary path 
export PATH=$PATH:~/go/bin

# If problem during install 
export GO111MODULE=auto 
go get -u github.com/rossigno1/gopatate

```

# Pre-requisite
None 

# How it works

It just brute force file and folder and give you the results...that's all but that's enough.
By default results are saved in **patalog.csv**.

```bash
# Simple brute force 
gopatate -u https://domain.com/FILE0 -f ~/Tools/SecLists/Discovery/Web/raft-large-directories.txt -x code=404

# Brute force with avoiding bad 200 (always the same size) and extension for discovering file
gopatate -u https://domain.com/FILE0 -f ~/Tools/SecLists/Discovery/Web/raft-large-directories.txt -x code=404,code=403,size=345 -ext php,ini,sql

# Brute force with avoiding bad 200 (where size less than 340) and extension for discovering file
gopatate -u https://domain.com/FILE0 -f ~/Tools/SecLists/Discovery/Web/raft-large-directories.txt -x code=404,code=403,size=340-infinite -ext php,ini,sql

# Brute force for finding your uploaded file
gopatate -u https://domain.com/upload/FILE0/myfile.phtm -f ~/Tools/SecLists/Discovery/Web/raft-large-directories.txt -x code=404,code=403,size=345-,msg="(.)*not found(.)*" 
```

For now it can only take one file. May be one day, i will take more files and integrate combo.

# No verified function 

The regexp function need to be tested.

Where size sup to : `size=345-infinite` 

# Need to be added soon 
- [ ] Add brute force in header 
- [ ] Add brute force in data 
