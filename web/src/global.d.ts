import type { TreantChart } from "./lib/convertToTreant";
export {};

declare global {
  interface Window {
    Treant: new (config: TreantChart) => void;
  }
}
