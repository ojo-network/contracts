import { ethers } from "hardhat";

async function main() {
  const Oracle = await ethers.getContractFactory("PriceFeed");
  const oracle = await Oracle.deploy();

  await oracle.deployed();

  console.log(
    `Oracle deployed to ${oracle.address}`
  );

  console.log(
    `deployer address ${(await ethers.getSigners())[0].address}`
  )
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
