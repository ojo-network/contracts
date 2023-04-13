import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "hardhat-abi-exporter";
import * as dotenv from 'dotenv';

dotenv.config()

const config: HardhatUserConfig = {
  solidity: {
    version: "0.8.9",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      }
    }
  },

  networks:{
    hardhat:{
      accounts: [
        {
          privateKey: "$PRIV_KEY",
          balance: "1000000000000000000000" // 1000 ETH
        },
      ]
    },

    nat:{
      url:"https://triton.api.nautchain.xyz",
      accounts:["$PRIV_KEY"]
    },

    polygon:{
      url:"https://polygon-mainnet.g.alchemy.com/v2/TrWHBdV8QcsNLsXTY4qGhcZ_h5dfoPDP",
      accounts:["$PRIV_KEY"],
      gasPrice: 166000000000, // 166 gwei in wei
      gas: 100
    }
  }
};

export default config;
