import { HardhatUserConfig, task } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "hardhat-abi-exporter";
import * as dotenv from 'dotenv';

dotenv.config();
let priv_key: string = process.env.PRIVATE_KEY || "";

function get_networks(){
    let networks: any = {};
    if(process.env.NETWORKS){
        let network_list = process.env.NETWORKS.split(",");
        network_list.forEach((network: string) => {
            networks[network] = {
                url: process.env[network + "_URL"] || "",
                accounts: [priv_key]
            }
        })
    }

    networks["hardhat"]={
         chainId:1,
      mining:{
        auto: true,
        interval:1000
      }
    }

    return networks;
}

const config: HardhatUserConfig = {
  solidity: {
    version: "0.8.9",
    settings: {
      optimizer: {
        enabled: true,
        runs: 800
      }
    }
  }
};

config.networks=get_networks();

export default config;
