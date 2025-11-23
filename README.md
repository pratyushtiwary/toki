<h1 align="center"><img src="./logo.svg" width="50" height="50"><br>Toki</h1>


<p align="center">Toki is an auth orchestrator written in Go. Toki simplifies token renewal through its config and use of signals.</p>

## 10000ft Overview

<img src="./toki_10000ft_overview.png" width="1000" height="1000">

## Usage

`pipeline-config.json`

```json
[
	{
		"command": "python3 test.py",
		"auth": [
			{
				"strategy": "custom",
				"params": {
					"expiry": 60,
					"refresh_after": 50,
					"command": "sleep 5 && echo 1"
				}
			}
		]
	},
	{
		"command": "python3 test.py",
		"auth": [
			{
				"strategy": "custom",
				"params": {
					"expiry": 60,
					"refresh_after": 50,
					"command": "sleep 5 && echo 1"
				}
			}
		]
	}
]
```

`command`: `toki pipeline-config.json`
