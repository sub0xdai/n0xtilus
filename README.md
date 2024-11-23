# n0xtilus

## Overview

CLI tool written in go, designed to trade perpetual swaps with trade execution and position-sizing automation.

## EXAMPLE:

_Freddy see chart. He think number go up on chart. Freddy just open tool and type trade. Tool ask Freddy which pair, entry and stop loss. Position size is then calculated and trade is placed. Freddy doesn't get greedy. The hard part is taken care of. Freddy is happy._

### Why?

- If you do not manage risk the market will manage it for you
- I want to manage the risk without using my brain everytime 

## Configuration

1. Copy the template config file:
   ```bash
   cp config.yaml.template config.yaml
   ```

2. Edit `config.yaml` with your API credentials:
   ```yaml
   api_key: "your_api_key"
   api_secret: "your_api_secret"
   api_base_url: "https://api.example.com"
   risk_percentage: 2
   test_mode: false
   ```

   > ⚠️ Never commit your `config.yaml` file! It's automatically ignored by `.gitignore`.

3. For testing without real API credentials, set `test_mode: true` in your config.

### Future Features:

1. Short and Long positions available
2. Ability to paste API from your favourite exchange
3. Advanced trade management 
4. Support for any pair available on your favoured exchange
