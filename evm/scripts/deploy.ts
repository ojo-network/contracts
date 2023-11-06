import { ethers } from "hardhat";

async function main() {
  const [deployer, relayer] = await ethers.getSigners();

  const Oracle = await ethers.getContractFactory("PriceFeed");
  const oracle = await Oracle.deploy(relayer.address);

  await oracle.deployed();

  console.log(
    `Oracle deployed to ${oracle.address}`
  );

  console.log(
    `deployer address ${deployer.address}`
  )

  console.log(
    `relayer address ${relayer.address}`
  )
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
