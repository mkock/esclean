import { usedFunctionFive } from "./dir1/file4";

export function unusedFunctionFour() {
  const calc = usedFunctionFive(10, 20, 30);

  const time = new Date();
  return calc + " The time is " + time.toISOString();
}

export function usedFunctionFour() {
  const time = new Date();
  return "The time is " + time.toLocaleDateString();
}
