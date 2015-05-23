## mouryou

#### Requirements

  - [Golang](https://golang.org/) >= 1
  - [Apache HTTP Server](http://httpd.apache.org/) with [mod_status](http://httpd.apache.org/docs/2.4/mod/mod_status.html)
  - [IPVS](http://www.linuxvirtualserver.org/software/ipvs.html)

#### Installation

    $ git clone git://github.com/hico-horiuchi/mouryou.git
    $ cd mouryou
    $ make gom
    $ sudo make install

#### Configuration

`~/.mouryou.json`

    {
      "cluster": {
        "lb": {
          "vip": "192.168.11.11",
          "algorithem": "wlc",
          "threshold": 0.8,
          "scaleout": 2,
          "scalein": 8
        },
        "hvs": [
          {
            "host": "192.168.11.20",
            "vms": [
              {
                "name": "web-server-1",
                "host": "192.168.11.21"
              }
            ]
          }
        ]
      },
      "timeout": 1,
      "wait": 30
    }
    
#### License

mouryou is released under the [MIT license](https://raw.githubusercontent.com/hico-horiuchi/mouryou/master/LICENSE).
