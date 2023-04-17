import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "hardhat-abi-exporter";
import * as dotenv from 'dotenv';

dotenv.config();
let priv_key: string = process.env.PRIVATE_KEY as string;

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
    hardhat:{},

    nat:{
      url:"https://triton.api.nautchain.xyz",
      accounts: [priv_key]
    },

    polygon:{
      url:"https://polygon-mainnet.g.alchemy.com/v2/TrWHBdV8QcsNLsXTY4qGhcZ_h5dfoPDP",
      accounts: [priv_key],
    }
  }
};

export default config;
