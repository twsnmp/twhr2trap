# twhr2trap
Send SNMP TRAP by monitoring host resources

If CPU usage, memory usage, load, or disk usage exceeds threshold values
Program to send SNMP TRAP

CPU使用率、メモリー使用率、負荷、ディスク使用率が閾値を超えたら
SNMPのTRAPを送信するプログラム

[![Godoc Reference](https://godoc.org/github.com/twsnmp/twhr2trap?status.svg)](http://godoc.org/github.com/twsnmp/twhr2trap)
[![Go Report Card](https://goreportcard.com/badge/twsnmp/twhr2trap)](https://goreportcard.com/report/twsnmp/twhr2rmon)

## Overview/概要

The standard SNMP agent does not have TRAP for host resources.
So, a simple sensor program that sends an SNMP TRAP or syslog when CPU usage, memory usage, load, or disk usage exceeds a set
I have created a simple sensor program that sends SNMP TRAP or syslog when CPU usage, memory usage, load, and disk usage exceed a set threshold.
I created a simple sensor program that sends an SNMP TRAP or syslog when CPU usage, memory usage, load, or disk usage exceeds a set threshold.
It monitors

- CPU usage
- Memory usage
- Load
- Disk usage


標準のSNMPエージェントには、ホストリソースのTRAPがありません。
そこで、CPU使用率、メモリー使用率、負荷、ディスク使用率が設定した
閾値を超えたらSNMPのTRAPかsyslogを送信するシンプルなセンサープログラムを
作ってみました。
モニタするのは

- CPU使用率
- メモリー使用率
- 負荷
- ディスク使用率

です。

## Status/ステータス

v1.0.0をリリースしました。(2023/1/21)  
（基本的な機能の動作する状態）  

## Build/ビルド方法

Build is done with make.
````
make
```

The following targets can be specified.
```
  all: Build all executables (optional)
  mac: Build executables for Mac
  clean: Delete the built executables
  zip: Create a zip file for release.
```

will create executables for MacOS, Windows, Linux(amd64), and Linux(arm) in the  
The executable files for MacOS, Windows, Linux(amd64), and Linux(arm) will be created in the ``dist`` directory.


To create a ZIP file for distribution, type
```
$make zip
```
will create a ZIP file in the ``dist/`` directory.


ビルドはmakeで行います。
```
$make
```
以下のターゲットが指定できます。
```
  all        全実行ファイルのビルド（省略可能）
  mac        Mac用の実行ファイルのビルド
  clean      ビルドした実行ファイルの削除
  zip        リリース用のZIPファイルを作成
```

```
$make
```
を実行すれば、MacOS,Windows,Linux(amd64),Linux(arm)用の実行ファイルが、  
`dist`のディレクトリに作成されます。


配布用のZIPファイルを作成するためには、
```
$make zip
```
を実行します。ZIPファイルが`dist/`ディレクトリに作成されます。

### Usage/使用方法

```
Usage of ./dist/twhr2trap.app:
  -all
    	send trap continuously
  -community string
    	snmp v2c  trap community
  -cpu int
    	cpu usage threshold 0=disable
  -disk int
    	disk usage threshold 0=disable
  -eid string
    	snmp v3 engine ID
  -interval int
    	check interval(sec) (default 10)
  -load int
    	load usage threshold 0=disable
  -mem int
    	memory usage threshold 0=disable
  -mode string
    	snmp trap mode (v2c|v3Auth|v3AuthPriv) (default "v2c")
  -password string
    	snmp v3 password
  -syslog string
    	syslog destnation list
  -trap string
    	trap destnation list
  -user string
    	snmp v3 user


```

Multiple trap and syslog destinations can be specified, separated by commas.  
You can also specify a port number followed by :.

trap,syslogの送信先はカンマ区切りで複数指定できます。  
:に続けてポート番号を指定することもできます。

```
-trap 192.168.1.1,192.168.1.2:8162
-syslog 192.168.1.1,192.168.1.2:5514
```


### Run/起動方法

To start, either a tarp destination (-trap) or a syslog destination (-syslog) is required.
If none of the CPU or other thresholds are specified, nothing will be sent.
(This is a useless operation.)

In Mac OS, Windows, and Linux environments, it can be started with the following command.  
(The example is for Linux)

```
#./twhr2trap  -trap 192.168.1.1 -syslog 192.168.1.1 -cpu 90 -mem 90 -load 20 -disk 90 -coummnity trap
```

起動するためには、tarpの送信先(-trap)かsyslogの送信先(-syslog)が必要です。
CPUなどのしきい値を１つも指定しなければ何も送信しません。
（無駄な動作をします。）

Mac OS,Windows,Linuxの環境では以下のコマンドで起動できます。  
（例はLinux場合）

```
#./twhr2trap  -trap 192.168.1.1 -syslog 192.168.1.1 -cpu 90 -mem 90 -load 20 -disk 90 -coummnity trap
```

## Package/TWSNMP FCのパッケージ

Include twhr2trap in the TWSNMP FC package.  
Windows/Mac OS/Linux(amd64,arm) are available.  
For more information, please visit  

https://note.com/twsnmp/n/nc6e49c284afb  

for more information.


TWSNMP FCのパッケージにtwhr2trapを含めます。  
Windows/Mac OS/Linux(amd64,arm)があります。  
詳しくは、  
https://note.com/twsnmp/n/nc6e49c284afb  
を見てください。

## Copyright

see ./LICENSE

```
Copyright 2023 Masayuki Yamai
```
