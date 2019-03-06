import {
  usedFunctionOne,
  usedFunctionTwo,
  usedFunctionThree,
  usedFunctionFour,
  usedFunctionFive
} from "./file1";
import { usedFunctionThree } from "./file2";

const greet = usedFunctionOne();
const bye = usedFunctionTwo();
const end = usedFunctionThree();

console.log(greet);
console.log(bye);
console.log(end);
