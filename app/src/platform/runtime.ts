import type { Snapshot } from './api';
import type { CaseSdk } from './sdk';
export type CaseContext = { initial: Snapshot; sdk: CaseSdk; exit: () => void };
export type CaseImplementation = {
  case_id: string; case_version: string;
  mount: (element: HTMLElement, context: CaseContext) => () => void;
};
