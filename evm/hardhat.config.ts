import { HardhatUserConfig } from 'hardhat/config';
import '@nomicfoundation/hardhat-toolbox';

require('dotenv').config();

const optimismGoerliUrl =
   process.env.ALCHEMY_API_KEY ?
      `https://opt-goerli.g.alchemy.com/v2/${process.env.ALCHEMY_API_KEY}` :
      process.env.OPTIMISM_GOERLI_URL

const config: HardhatUserConfig = {
  solidity: {
    version: '0.8.17',
  },
  networks: {
    // for mainnet
    'base-mainnet': {
      url: 'https://mainnet.base.org',
      accounts: [process.env.WALLET_KEY as string],
      gasPrice: 1000000000,
    },
    // for testnet
    'base-goerli': {
      url: 'https://goerli.base.org',
      accounts: [process.env.WALLET_KEY as string],
      gasPrice: 1000000000,
    },
    // for local dev environment
    'base-local': {
      url: 'http://localhost:8545',
      accounts: [process.env.WALLET_KEY as string],
      gasPrice: 1000000000,
    },
    "optimism-goerli": {
      url: optimismGoerliUrl,
      accounts: { mnemonic: process.env.MNEMONIC },
      gasPrice: 1000000000,
   }
  },
  defaultNetwork: 'hardhat',

  etherscan: {
    apiKey: {
     "base-goerli": "PLACEHOLDER_STRING",
     "optimisticGoerli": "2H6U6NDAI9GXTKSJM9TBEMVRB5ZPQ4E9JZ",
    },
    customChains: [
      {
        network: "base-goerli",
        chainId: 84531,
        urls: {
         apiURL: "https://api-goerli.basescan.org/api",
         browserURL: "https://goerli.basescan.org"
        }
      }
    ]
  },

};

export default config;
