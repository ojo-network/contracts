import { time, loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { anyValue } from "@nomicfoundation/hardhat-chai-matchers/withArgs";
import { expect } from "chai";
import { ethers } from "hardhat";

interface Data{
    id: number
    assetName: string,
    value: number,
    deviation: number,
    resolveTime:number,
    medians: number[],
}


const generateData=function(n:number, id:number): Data[]{
    const dataArr: Data[] = [];
    for(let i=0;i<n;i++){
        const name = ethers.utils.formatBytes32String(`asset${i}`)
        const  value = Math.floor(Math.random()*10**9)
        const deviation = Math.floor(Math.random()*10**9)
        const resolveTime = Math.floor(Date.now()/1000)
        const medians: number[] = new Array(10).fill(0);

        for (let j=0;j<10;j++){
            medians[j]=(Math.random()*10**9)
        }
        let data={id:id,assetName:name,value:value,deviation:deviation,medians:medians,resolveTime:resolveTime}
        dataArr.push(data);
    }
    return dataArr;
}


describe("Deploy", function () {
    async function deployOracle() {
      // Contracts are deployed using the first signer/account by default
      const [deployer, otherAdmin, relayer, otherRelayer] = await ethers.getSigners();
   
      const Oracle = await ethers.getContractFactory("PriceFeed");
      const oracle = await Oracle.deploy();
      const relayerRole = await oracle.RELAYER_ROLE();
      const defaultRole = await oracle.DEFAULT_ADMIN_ROLE();

      return {oracle, deployer, otherAdmin, relayer, otherRelayer, relayerRole, defaultRole}
    }


    it("check admin, ownership transfer and relayer address", async  function(){
        const {oracle, deployer,otherAdmin, relayer , relayerRole, defaultRole} = await loadFixture(deployOracle);
        
        // check owners
        expect(await oracle.owner()).eq(deployer.address)
        
        // check roles
        expect(await oracle.hasRole(relayerRole, deployer.address)).eq(true)
        expect(await oracle.hasRole(defaultRole, deployer.address)).eq(true)

        expect(await oracle.hasRole(relayerRole, otherAdmin.address)).eq(false)
        expect(await oracle.hasRole(defaultRole, relayer.address)).eq(false)

        // transfer ownership
        await oracle.transferOwnership(otherAdmin.address)
        expect(await oracle.owner()).not.eq(otherAdmin.address)

        // claim ownership
        
        // false temp owner
        expect(oracle.connect(relayer).claimOwnership()).to.be.revertedWithPanic;
        
        // claim ownership transfer
        await oracle.connect(otherAdmin).claimOwnership()
        expect(await oracle.owner()).eq(otherAdmin.address)

        // check for old roles
        expect(await oracle.hasRole(relayerRole, deployer.address)).eq(false)
        expect(await oracle.hasRole(defaultRole, deployer.address)).eq(false)

        expect(await oracle.hasRole(relayerRole, otherAdmin.address)).eq(true)
        expect(await oracle.hasRole(defaultRole, otherAdmin.address)).eq(true)
    })

    it("relayer access control", async  function(){
        const {oracle, otherAdmin, relayer, relayerRole} = await loadFixture(deployOracle);
     
        // grant role with non admin 
        expect(oracle.connect(otherAdmin).grantRole(relayerRole, relayer.address));
        
        // grant role with admin
        await oracle.grantRole(relayerRole, relayer.address)
        expect(await oracle.hasRole(relayerRole, relayer.address)).eq(true)

        // revoke role with non admin
        expect(oracle.connect(otherAdmin).revokeRole(relayerRole, relayer.address)).to.be.revertedWithCustomError;

        // revoke role with admin
        await oracle.revokeRole(relayerRole, relayer.address)
        expect(await oracle.hasRole(relayerRole, relayer.address)).eq(false)
    })

    it("post price, deviation and median data",async () => {
        
        const{oracle, otherAdmin,relayerRole,relayer} = await loadFixture(deployOracle);

        // grant role with admin
        await oracle.grantRole(relayerRole, relayer.address)

        let data= generateData(10,1)
        
        await oracle.connect(relayer).postPrices(data)

        let priceValue=await oracle.getPriceData(data[0].assetName)

        let allData= await oracle.getPriceDataBulk(data.map(d=>d.assetName))
        expect(allData.length).eq(data.length)
        
        for (let i=0;i<data.length;i++){
            expect(allData[i].value).eq(data[i].value)
            expect(allData[i].id).eq(data[i].id)
            expect(allData[i].resolveTime).eq(data[i].resolveTime)
            expect(allData[i].assetName).eq(data[i].assetName)
        }
    })
})