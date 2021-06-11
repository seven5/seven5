import { inputPolicyBase } from './inputpolicy';
import * as S5Event from '../event';
import dispatchAgent from './dispatchagent';
import * as S5Interactor from '../interactor';

export default class positionalPolicy extends inputPolicyBase {
  root: S5Interactor.Root;

  constructor(root: S5Interactor.Root) {
    super();
    this.root = root;
  }

  // processing events positionally means doing a pick
  ProcessEvent(e: S5Event.Event): boolean {
    let found = false;
    if (e.PickList === null) {
      const pl = new S5Event.PickList();
      pl.Build(e, this.root);
      e.PickList = pl;
    }

    this.agent.forEach((agent: dispatchAgent): void => {
      if (!found) {
        found = agent.Dispatch(e);
      }
    });
    return found;
  }
}
