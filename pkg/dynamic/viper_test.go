package dynamic

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// any approach to require this configuration into your program.
var yamlExample = []byte(`
JWTPublicKey:
- name: "dev-vault-active"
  comment: "copied from VEGA_JWT_PUBLIC_KEY"
  not_before: 2020-01-01
  not_after: 2022-01-01
  key: |
    -----BEGIN PUBLIC KEY-----
    MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAyqw4d1SSY6fk61jc4B70
    xEaJb3h4gczV6vmy2GS4+cizBCC4bjuRrL72/rlVUBszHFLar40nPOAlrGD8ZrMY
    Vn/lEvCHL19r1tHs/iHU9wdX6nY6Pkkd7FPnosFh6uB80KdQV1ahC5wQ60/0yA+O
    CiEbj8YHj/K9y2BwV4G6+FFda7URh9P9zDnzZ5uwYx9FXOxmNOWIo3yjw2goyUJw
    2s90Zlce14k18Uo1/wKjtMw5girxbi3tl8pQqm3c2AHSllmfNAyW2hOTDrT5M3rA
    gktYGTCdJDv9sYZkE10P7o2jx39yP8314Da25RfTSK9Og9UGv8NbYOjBCo8/UbAp
    t7h4kbxMcOvVpf2PmSIKC3859y5lBmH5y57YwuejpcPv/QRPoO6MhzXfS2LsjQIJ
    9XtqvInhwosZy9AId7mL804o7Dzsx+EZtCcw//P36ZO1wuTRomSUaBOL5G+M6Poy
    riosmwtPDJj7zcr6RlM8B4RDPrz9Q9/s4GGIgL0zV0GFXXKpPQe9v1J6iG5zRIcW
    61RHWlaMT5Op6kTLZTaiB1qAqNZi9ljKtPXN2vAO0Lq03MCQNmdgqmbghpKxGg9u
    /17+tYVb5MAhwHcqXF0r1rMHXb69y1bwRFnrCaHF/WFzZJBmOhanhs7OyPdpJiHz
    t9mxhZZUU8FCb9p3MyRvg3ECAwEAAQ==
    -----END PUBLIC KEY-----
- name: "dev-vault-previous"
  comment: "copied from somewhere"
  not_before: 2020-01-01
  not_after: 2022-01-01
  key: |
    -----BEGIN PUBLIC KEY-----
    MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAl+J2z++v3JcSL2r6Kfo9
    9eeoEXbksQvHSjKYSR6h3V4y3T6ZCBo5Fd6BMWmyMuvkj+FGzjOydBbYFr7oYJih
    5GG8GZFHHiPe3YOtlst0ccG0MKN1Xyqu73OlwShtopqf77dGfqcxBerIkXuRr7R8
    3KwCohPla7dhrCpSeTvXcdMHoERXpeSY0wXKpAc1GtXu8G0cMuHzPwLPUQhYAwUg
    pSH1D1JKsQVsN8BW58NXDhuVuBVOsOBzO///dQT35dXwFJa7v57CqncJp2JjpIwT
    NaWwjWcWBXvCoUonFS55pXmJo2+NAOHRMEnnEvpzOQLTwnBW3xffRlw96/x5BTH7
    adXmcnE1UM8JQ7lGnbOFdsyfN6GvLoITZA5YVyOX/hfHeZOQPF1fGcvsANqwL7ZJ
    CHmgnJMD2X4X1njv6lftXFDJFqN/Z/a/zTLgLSKsZrl44zwbjts4leCs/mOhKDwt
    3QtUCFQG8pkVBbbtAWx7tlL18H9N23p67aL5OQi+ZJPs3NH538EV/SFF2aWrLRqX
    hwdOGFNkuBrH/qYPMB41kzHy7c7JhVf+zcedPCnXC1YwD1nSJtKDIbIfmTaIMMaU
    vvVhw9eR1/TdBGUcnDBQA5FZWBvGL7q0QwNbSv/ne9Y+QuxqWMAyUimXYvc0LLuL
    KLPw52VL0SvM9rHnW+sqIJ8CAwEAAQ==
    -----END PUBLIC KEY-----
- name: "integrationTest"
  comment: "copied from this repo: /test/config"
  not_before: 2020-01-01
  not_after: 2022-01-01
  key: |
    -----BEGIN PUBLIC KEY-----
    MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA2ljY82CRfjTQ9eZaDLO/
    W02/eGwdlzAvf+54dXqDd5AjxEtNtIDNGzuQiH+74zMOlGJbH6bh74BgRbXmTyP4
    dgbJkfdXtiLo161wJTptEhUBDF3+evCUZZ4+M8xCrV97lZXId6Lm9yDO+FSEMHbM
    3XdA07vVSldvr/feQHKuGBrcBLHzBjme4s+TKFD6+nEhp9WujPfw6SgpHGZCA3HN
    6KwiWWJ9TDs2PiA0vgoW3DT8FHhu0hDkcusYImcdc2AUVMSXLqbFMAc7fJzKwuEr
    CAI4wOmg4C+7XczSuDkEhcZLhGOxxDNe945ExiTqQeLn29uJ1T6LbwfPy2IS16Ji
    Pxa/lYqZc1C9+s+8iFOYb/rEzzyfmSI6Jb7SVFgq2qJftgX887Wr9xi5NFWd8TsA
    bws92P/2Gge6Mfyr5CVs3nI4wcMkbiBznPsJWuZnsbjAENlGXLBE7ljyygAwuB/A
    ysDobMKjenG8552q93OUwVDuGyAi3zF0nEvFaToSszoEGIZysRbKjW1INPWnTFcA
    +U85o8kj+Drr8DTohBrZ3G86bg1YHpTvMnUQuxZ+usLxynGsEckjWz+9QevYt1s4
    6ZaSfU4L9JM5USRFX3RpoSf5MdGhb2G8qWNleqXdhTXEHpeUFZcdRfASKxgMWJ+x
    YflMznsWkLa8MG5Z0kqhYu0CAwEAAQ==
    -----END PUBLIC KEY-----

`)

func TestBasicViper(t *testing.T) {
	vc := NewViperConfig(ConfigFormat("yaml"),
		ConfigSource(bytes.NewBuffer(yamlExample)),
		AutoBootstrap(false),
		WatchChanges(false),
	)
	err := vc.Bootstrap()
	assert.NoError(t, err)

	arr, err := vc.PublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(arr))
}
