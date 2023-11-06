import { ethers, upgrades } from "hardhat";

async function main() {
  const [deployer] = await ethers.getSigners();

  const Oracle = await ethers.getContractFactory("PriceFeed");
  const oracle = await upgrades.upgradeProxy("address", Oracle)

  await oracle.deployed();

  console.log(
    `Oracle deployed to ${oracle.address}`
  );

  console.log(
    `deployer address ${deployer.address}`
  )
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
