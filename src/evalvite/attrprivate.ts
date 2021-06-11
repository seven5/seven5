import attrPrivate, { topoMarkValues } from './base';
import attribute, { updateable } from './attribute';
import * as S5Conf from '../conf';
import * as S5Err from '../error';
import { decodeTypename, decodeValue } from './decode';

// do this intermediate class so we can share code with the dirty marks and the outgoing edges
export default abstract class attrPrivateImpl<T> implements attrPrivate<T> {
  outEdges: Array<attrPrivate<unknown>>; // people that depend on my value

  inEdges: Array<attrPrivate<unknown>>; // people whose value I consume

  dirty: boolean;

  boundUpdateable = new Array<updateable<T>>();

  topoMark = topoMarkValues.none;

  _eager = false;

  /**
   * Debug name.
   */
  dName: string;

  constructor(n: string) {
    this.dName = n;
    this.outEdges = [] as Array<attrPrivate<unknown>>;
    // basic attribute uses inEdges to keep track of function inputs
    this.inEdges = [] as Array<attrPrivate<unknown>>;

    // this is not expected... it is correct to do this because in
    // any case where anything changes, you'll get correct dirty
    // marks... and a demand on most nodes in the start state is
    // fine (simple,array)
    this.dirty = false;
  }

  // returns -1 if the target is not present
  findBoundUpdaterIndex(target: updateable<T>): number {
    for (let i = 0; i < this.boundUpdateable.length; i += 1) {
      if (this.boundUpdateable[i] === target) {
        return i;
      }
    }
    return -1;
  }

  mark(): topoMarkValues {
    return this.topoMark;
  }

  setMark(t: topoMarkValues): void {
    this.topoMark = t;
  }

  DebugName = (): string => {
    return `${this.dName}:${this.attributeTypename()}`;
  };

  AddUpdateable(c: updateable<T>): void {
    if (!c) {
      if (S5Conf.Vars.Debug) {
        S5Conf.Log(
          `EVDEBUG: ${this.DebugName()} ignoring attempt to addComponent() with a falsey value`
        );
      }
      return;
    }
    if (this.findBoundUpdaterIndex(c) >= 0) {
      if (S5Conf.Vars.Debug) {
        S5Conf.Log(
          `EVDEBUG: ${this.DebugName()} ignoring attempt to addComponent() which is already present`
        );
      }
      return;
    }
    if (S5Conf.Vars.Debug) {
      S5Conf.Log(`EVDEBUG: ${this.DebugName()}: binding react component`);
    }
    this.boundUpdateable.push(c);
    this.markDirty();
    c.ConnectAttribute(this);
    c.UpdateValue(this, this.Get());
  }

  RemoveUpdateable(c: updateable<T>): void {
    if (!c) {
      if (S5Conf.Vars.Debug) {
        S5Conf.Log(
          `EVDEBUG: ${this.DebugName()} ignoring attempt to removeComponent() with a falsey value`
        );
      }
      return;
    }
    const index = this.findBoundUpdaterIndex(c);
    if (index < 0) {
      if (S5Conf.Vars.Debug) {
        S5Conf.Log(
          `EVDEBUG: ${this.DebugName()} ignoring attempt to removeComponent() which is not present`
        );
      }
      return;
    }
    const target = this.boundUpdateable[index];

    if (this.boundUpdateable.length === 1) {
      this.boundUpdateable = new Array<updateable<T>>();
      target.DisconnectAttribute(this);
      return;
    }
    // this maneuver is safe, the list is not ordered
    const tmpComponent = this.boundUpdateable[0];
    this.boundUpdateable[0] = this.boundUpdateable[index];
    this.boundUpdateable[index] = tmpComponent;
    this.boundUpdateable.pop();
    target.DisconnectAttribute(this);
  }

  // note that calling addOutgoing does NOT mark dirty, that's the caller's
  // responsibility
  addOutgoing(target: attrPrivate<unknown>): void {
    if (S5Conf.Vars.Debug) {
      S5Conf.Log(
        `EVDEBUG: ${this.DebugName()}: adding edge -> ${target.DebugName()}, now ${
          this.outEdges.length + 1
        } edges`
      );
    }
    let found = false;
    for (let i = 0; i < this.outEdges.length; i += 1) {
      if (this.outEdges[i] === target) {
        found = true;
        break;
      }
    }
    if (found) {
      S5Err.NewError(
        S5Err.Messages.DuplicateOutgoing,
        `from ${target.DebugName()} to ${this.DebugName()}`
      );
    }
    this.outEdges.push(target);
  }

  getOutgoing(): Array<attrPrivate<unknown>> {
    return this.outEdges;
  }

  removeOutgoing = (
    target: attrPrivate<unknown>,
    allowMissing?: boolean
  ): void => {
    if (S5Conf.Vars.Debug) {
      S5Conf.Log(
        `EVDEBUG: ${this.DebugName()}: removing edge -> ${target.DebugName()}`
      );
    }
    let found = false;
    let index = -1; // to force throw

    for (let i = 0; i < this.outEdges.length; i += 1) {
      if (this.outEdges[i] === target) {
        found = true;
        index = i;
        break;
      }
    }
    if (!found) {
      if (!allowMissing) {
        throw new Error(
          `unable to find target in outgoing edges of ${this.DebugName()}`
        );
      }
    }

    if (this.outEdges.length > 1) {
      // the list is unordered so this trick is fine...there must be a bug in eslint
      // because  I don't see how to use destructuring more than I already am...
      // eslint-disable-next-line prefer-destructuring
      this.outEdges[index] = this.outEdges[0];
    }
    // contents pointed to by removed is now "dirty" in the sense
    // it changed (because it's sources are diff)
    const removed = this.outEdges.shift();
    if (removed !== undefined) {
      removed.markDirty();
    }
    this.markDirty();
  };

  addIncoming(attr: attrPrivate<unknown>): void {
    this.inEdges.push(attr);
  }

  removeIncoming(attr: attrPrivate<unknown>, allowMissing?: boolean): void {
    let found = false;
    let count = 0;
    this.inEdges.forEach((e) => {
      if (e === attr) {
        if (this.inEdges.length === 1) {
          // simple case
          this.inEdges = [];
        } else {
          // I really don't like this array destructuring notation, but eslint does
          [this.inEdges[count], this.inEdges[0]] = [
            this.inEdges[0],
            this.inEdges[count],
          ];
          this.inEdges.pop();
          found = true;
        }
      }
      count += 1;
    });
    if (!found) {
      if (!allowMissing) {
        throw S5Err.NewError(
          S5Err.Messages.BadState,
          `${this.DebugName} not found in inEdges`
        );
      }
    }
  }

  markDirty(): void {
    if (this.dirty) {
      if (S5Conf.Vars.Debug) {
        S5Conf.Log(
          `EVDEBUG: ${this.DebugName()}: no reason to mark dirty, already dirty[${
            this.outEdges.length
          }],${this.outEdges.map((e) => e.DebugName())}`
        );
        S5Conf.Log(
          `${this.outEdges.map((e) =>
            (e as attrPrivateImpl<unknown>).dirty ? 'true' : 'false'
          )}`
        );
      }
      return;
    }
    if (S5Conf.Vars.Debug) {
      S5Conf.Log(`EVDEBUG: ${this.DebugName()}: beginning topo sort`);
    }
    const topoResult = attrPrivateImpl.topoSort(this);
    topoResult.forEach((a: attrPrivate<unknown>): void => {
      // xxx fixme(iansmith) should this be done via some API and no cast?
      (a as attrPrivateImpl<unknown>).dirty = true;
    });

    if (S5Conf.Vars.Debug) {
      S5Conf.Log(`EVDEBUG: ${this.DebugName()}: beginning eager pass`);
    }
    topoResult.forEach((a: attrPrivate<unknown>): void => {
      if (a.Eager) {
        a.Get();
      }
    });

    this.dirty = true;
    this.update();
  }

  markDirtyXXX(): void {
    if (this.dirty) {
      if (S5Conf.Vars.Debug) {
        S5Conf.Log(
          `EVDEBUG: ${this.DebugName()}: no reason to mark dirty, already dirty[${
            this.outEdges.length
          }],${this.outEdges.map((e) => e.DebugName())}`
        );
        S5Conf.Log(
          `${this.outEdges.map((e) =>
            (e as attrPrivateImpl<unknown>).dirty ? 'true' : 'false'
          )}`
        );
      }
      return;
    }
    if (S5Conf.Vars.Debug) {
      S5Conf.Log(
        `EVDEBUG: ${this.DebugName()}: marking dirty recursive start [${
          this.outEdges.length
        } edges] ------------- `
      );
    }
    let ct = 0;
    this.outEdges.forEach((n: attrPrivate<unknown>) => {
      if (S5Conf.Vars.Debug) {
        S5Conf.Log(
          `EVDEBUG: ${this.DebugName()}: recursion ${ct} -> ${n.DebugName()}`
        );
      }
      n.markDirty();
      ct += 1;
    });
    if (S5Conf.Vars.Debug) {
      S5Conf.Log(
        `EVDEBUG: ${this.DebugName()}: ------------- marking dirty recursive end`
      );
    }
    this.dirty = true;
    this.update();
  }

  update(newValue?: T, oldValue?: T): void {
    if (this.boundUpdateable.length > 0) {
      if (S5Conf.Vars.Debug) {
        S5Conf.Log(
          `EVDEBUG: ${this.DebugName()}: setting state in all interactors[${
            this.boundUpdateable.length
          }]`
        );
      }
      for (let i = 0; i < this.boundUpdateable.length; i += 1) {
        const comp = this.boundUpdateable[i];
        if (oldValue !== undefined) {
          comp.UpdateValue(this, newValue, oldValue);
        } else if (newValue !== undefined) {
          comp.UpdateValue(this, newValue);
        } else {
          comp.UpdateValue(this);
        }
      }
    } else if (
      S5Conf.Vars.WarnOnUnboundAttributes &&
      this.outEdges.length === 0
    ) {
      S5Conf.Warn(
        this.DebugName(),
        ' marked component dirty, but no connected component(s)'
      );
    }
  }

  // use provided impl in index.ts
  attributeTypename(): string {
    return decodeTypename(this);
  }

  // topological sort credited to Tarjan, 1976
  static topoSort = (start: attrPrivate<unknown>): attrPrivate<unknown>[] => {
    const topoResult = [] as attrPrivate<unknown>[];
    let temps = [] as attrPrivate<unknown>[];

    const visit = (n: attrPrivate<unknown>) => {
      if (n.mark() === topoMarkValues.Permanent) {
        return;
      }
      if (n.mark() === topoMarkValues.temporary) {
        temps.forEach((t: attrPrivate<unknown>) => {
          S5Conf.Log(t);
        });
        if (S5Conf.Vars.AbortOnCycle) {
          throw new Error(
            `Cycle detected in the dependency graph. Cycle involves ${n.DebugName()}`
          );
        } else {
          S5Conf.Warn('Cycle detected in dependency graph, breaking cycle...');
        }
        return;
      }
      n.setMark(topoMarkValues.temporary);
      temps.push(n);
      n.getOutgoing().forEach((m: attrPrivate<unknown>) => {
        visit(m);
      });
      // remove n from list
      temps = temps.filter((e: attrPrivate<unknown>) => {
        return e !== n;
      });
      n.setMark(topoMarkValues.Permanent);
      topoResult.unshift(n);
    };

    visit(start);
    if (temps.length > 0) {
      throw new Error(
        `Topological sort algorithm did not terminate correctly: ${temps.length} marks remaining`
      );
    }
    topoResult.forEach((n) => {
      n.setMark(topoMarkValues.none);
    }); // clean up for next time
    return topoResult;
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  decode(): any {
    const a = this.Get();
    return decodeValue(a);
  }

  typename(): string {
    return `${this.AttributeName()}`;
  }

  // eslint-disable-next-line class-methods-use-this
  AllowsSet(): boolean {
    return false;
  }

  // cleanAndNotify is slightly tricky.  We use the API method "markDirty"
  // which does the efficient marking of dependencies and dealing with
  // eager attributes that are dependent on us.  *AFTER* that, then we
  // check if we are eager and update if so, then we finish by marking
  // ourselves clean.
  cleanAndNotify(): void {
    this.markDirty();
    this.dirty = false; // we are now clean
  }

  get Eager(): boolean {
    return this._eager;
  }

  set Eager(b: boolean) {
    this._eager = b;
  }

  DropAttributeConnection(...other: attribute<unknown>[]): void {
    other.forEach((cand) => {
      this.removeOutgoing(cand as attrPrivateImpl<unknown>, true);
      (cand as attrPrivateImpl<unknown>).removeIncoming(this, true);
    });
  }

  abstract SetInputsAndFunc(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    inputs: Array<attribute<any>>,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    fn: (...a2: any[]) => T
  ): void;

  // child classes have to implement these three
  abstract AttributeName(): string;

  abstract Get(): T;

  abstract Set(v: T): void;
}
