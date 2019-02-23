import { usedFunctionFour } from "./file3";

export function usedFunctionThree() {
  const msg = "The time is " + usedFunctionFour() + ".\nBye, now";
  return msg;
}

export function unusedFunctionOne(name) {
  return "Hello, " + name + "!";
}

// This const is deprecated...
export const unusedGreeting = "Hello, anonymous!";
