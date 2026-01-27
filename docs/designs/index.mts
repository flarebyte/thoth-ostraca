import { ComponentCall } from "./common.mts";

const calls: ComponentCall[] = [];

const M = 1000000;
const K = 1000;

const cliArgs = ()=> {
   const searchargs: ComponentCall = {
       name: "serach",
       title: "Parse CLI args",
       directory: "cmd",
       mustHave: ["Use cobra lib"],
       processingTime: {
           minMilli: 1,
           maxMilli: 1
       }
   }
}