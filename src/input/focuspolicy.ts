import inputPolicy, { inputPolicyBase } from './inputpolicy';
import * as S5Event from '../event';

export default class focusPolicy
  extends inputPolicyBase
  implements inputPolicy {
  // eslint-disable-next-line class-methods-use-this,@typescript-eslint/no-unused-vars
  ProcessEvent(_: S5Event.Event): boolean {
    return false;
  }
}
