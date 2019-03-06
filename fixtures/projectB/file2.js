import { usedFunctionFour } from "./file3";
import { usedFunctionFive } from "./dir1/file4";

export function usedFunctionThree() {
  const msg = "The time is " + usedFunctionFour() + ".\nBye, now";
  return msg;
}

export function unusedFunctionOne(name) {
  usedFunctionFive();
  return "Hello, " + name + "!";
}

// This const is deprecated...
export const unusedGreeting = "Hello, anonymous!";
