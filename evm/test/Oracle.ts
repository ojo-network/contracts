import { time, loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";

interface Data{
    id: number
    assetName: string,
    value: number,
    deviation: number,
    resolveTime:any,
    values: number[],
}


const generateData=function(n:number, id:number, resolveTime:any): Data[]{
    const dataArr: Data[] = [];
    for(let i=0;i<n;i++){
        const name = ethers.utils.formatBytes32String(`asset${i}`)
        const value = Math.floor(Math.random()*10**9)
        const deviation = Math.floor(Math.random()*10**9)
        const medians: number[] = new Array(10).fill(0);

        for (let j=0;j<10;j++){
            medians[j]=Math.floor((Math.random()*10**9))
        }
        let data={id:id,assetName:name,value:value,deviation:deviation,values:medians,resolveTime:resolveTime}
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

        let block=await ethers.provider.getBlock("latest")
        let resolveTime = (block.timestamp+1000).toString()
        let data= generateData(10,1,resolveTime)
        
        // post and check price data
        await oracle.connect(relayer).postPrices(data,false)
        let allPrices= await oracle.getPriceDataBulk(data.map(d=>d.assetName))
        expect(allPrices.length).eq(data.length)
        
        for (let i=0;i<data.length;i++){
            expect(allPrices[i].value).eq(data[i].value)
            expect(allPrices[i].id).eq(data[i].id)
            expect(allPrices[i].resolveTime.toString()).eq(data[i].resolveTime)
            expect(allPrices[i].assetName).eq(data[i].assetName)
        }

        // post and check deviation data
        await oracle.connect(relayer).postDeviations(data,false)
        let allDeviations=await oracle.getDeviationDataBulk(data.map(d=>d.assetName))
        expect(allDeviations.length).eq(data.length)

        for (let i=0;i<allDeviations.length;i++){
            expect(allDeviations[i].value).eq(data[i].value)
            expect(allDeviations[i].id).eq(data[i].id)
            expect(allDeviations[i].resolveTime.toString()).eq(data[i].resolveTime)
            expect(allDeviations[i].assetName).eq(data[i].assetName)
        }

        // post and check median data
        await oracle.connect(relayer).postMedians(data,false)
        let allMedians=await oracle.getMedianDataBulk(data.map(d=>d.assetName))
        expect(allMedians.length).eq(data.length)        

        for (let i=0;i<allMedians.length;i++){        
            expect(allMedians[i][3]).deep.eq(data[i].values)
            expect(allMedians[i].id).eq(data[i].id)
            expect(allMedians[i].resolveTime.toString()).eq(data[i].resolveTime)
            expect(allMedians[i].assetName).eq(data[i].assetName)
        }
    })


    it("whitelist list check",async () => {
        const{oracle, otherAdmin,relayerRole,relayer, otherRelayer} = await loadFixture(deployOracle);

        // grant role with admin
        await oracle.grantRole(relayerRole, relayer.address)

        let block=await ethers.provider.getBlock("latest")
        let resolveTime = (block.timestamp+1000).toString()
        let data= generateData(10,1,resolveTime)
        
        // post and check price data
        await oracle.connect(relayer).postPrices(data,false)

        // enable whitelist
        await oracle.setWhitelistStatus(true);
        await oracle.whitelistAddress(otherAdmin.address);

        // query using non whitelisted account
        expect(oracle.connect(otherRelayer).getPriceDataBulk(data.map(d=>d.assetName))).to.be.revertedWithCustomError

        // query using whitelisted account 
        let prices=await oracle.connect(otherAdmin.address).getPriceDataBulk(data.map(d=>d.assetName))
        expect(prices.length).eq(data.length)
    })

    it("median status check",async()=>{
        const{oracle, otherAdmin,relayerRole,relayer, otherRelayer} = await loadFixture(deployOracle);

        // grant role with admin
        await oracle.grantRole(relayerRole, relayer.address)

        let block=await ethers.provider.getBlock("latest")
        let resolveTime = (block.timestamp+1000).toString()
        let data= generateData(10,1,resolveTime)
        
        // post median data
        await oracle.connect(relayer).postMedians(data,false)

        // disable median query
        await oracle.setMedianStatus(true);
        expect(oracle.getMedianDataBulk(data.map(d=>d.assetName))).to.be.revertedWithCustomError

        await oracle.setMedianStatus(false)
        await oracle.getMedianDataBulk(data.map(d=>d.assetName))
    })
})