# tz-golang-graphite

1. Install Golang
  
  https://golang.org/dl/  
  wget https://storage.googleapis.com/golang/go1.5.1.darwin-amd64.pkg  
  cd  
  echo '' >> .bash_profile  
  echo 'export PATH=$PATH:/usr/local/go/bin' >> .bash_profile  
  source .bash_profile  

2. install glide
	https://github.com/Masterminds/glide

3. Get libraries
  
  export GOPATH=/Users/dhong/Documents/workspace/go
  mkdir $GOPATH  
  cd $GOPATH  
  cd src/tz.com/tz-golang-graphite

  glide up

3. run IDE

3-1 golang intellij
  
  1. get intellij  
    https://www.jetbrains.com/idea/  
  2. Preference > Plugins > Browse repositories > Manage repositories > add  
    https://plugins.jetbrains.com/plugins/alpha/5047  
    Select "Go" > install plugin > restart  
  3. context menu > open module settings  
    > Project > Project SDK > Go 1.5.1  
    > Platform Setttings > SDKs, add "go sdk"   
      "/usr/local/go"  
    > Project Setttings > Libraries, add "java"  
      select all folders in /Users/dhong/Documents/go/src > Classes  
    > Preference > Go Libraries >  
      Global Libraries > /Users/dhong/Documents/go  
      Project Libraries > /Users/dhong/Documents/workspace/dhc4  
  4. Change the module type to Go project  

3-2 golang eclipse
  
  1. install Eclipse IDE for Java EE Developers  
    https://www.eclipse.org/downloads/?osType=macosx  
  2. install jdk8  
    http://www.oracle.com/technetwork/java/javase/downloads/jdk8-downloads-2133151.html  
  3. plugin installation (goclipse)  
    add repository: http://goclipse.github.io/releases/  
    Select goclipse in available software  
  4. install Eclipse CDT (C/C++ Development Tooling)   
  5. install gdb  
    brew install gdb  
    export PATH=/usr/local/bin:$PATH  
  6. preferences in eclipse  
    GOROOT: /usr/local/go  
    GOOS: darwin  
    GOARCH: amd64  
    GOPATH: /Users/dhong/Documents/go  
  7. run eclipse with sudo  
    cd /Applications/Eclipse.app/Contents/MacOS  
    export GOPATH=/Users/dhong/Documents/go  
    export PKG_CONFIG_PATH=/usr/share/pkgconfig/lib/pkgconfig  
    sudo eclipse  
    
    
