export function usedFunctionOne() {
  return "Hello";
}

export function usedFunctionTwo() {
  return "Goodbye";
}

export function usedFunctionThree() {
  return usedFunctionTwo();
}

export function usedFunctionFour() {
  return usedFunctionThree();
}

export function usedFunctionFive() {
  return usedFunctionFour();
}

export function unusedFunctionOne() {
  return usedFunctionFour();
}
