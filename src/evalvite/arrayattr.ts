import attrPrivate from './base';
import attrPrivateImpl from './attrprivate';
import attribute from './attribute';

import { modelToAttrFields } from './typeutils';
import { decodeValue } from './decode';
import * as S5Conf from '../conf';
import * as S5Err from '../error';

let { IdCounter } = S5Conf.Vars;

export default class arrayAttribute<
  T extends Record<string, unknown>
> extends attrPrivateImpl<T[]> {
  inner: T[] = [] as T[];

  constructor(debugName?: string) {
    super(debugName || `[array attribute ${IdCounter}]`);
    IdCounter += 1;
  }

  attributeTypename(): string {
    return super.attributeTypename();
  }

  static iterateAttrs(
    item: Record<string, unknown>,
    fn: (attr: attrPrivate<unknown>) => void
  ): void {
    // we turn off eslint below because the keys that result here are already
    // checked in the modelToAttrFields() function.
    const keys = modelToAttrFields(item);
    keys.forEach((k: string) => {
      if (k.indexOf('$') < 0) {
        const prop = item[k] as attrPrivate<unknown>;
        fn(prop);
      }
    });
  }

  // xxx fixme(iansmith) all of the "standard" methods like push,pop,length
  // xxx are lower cased to make them the "same" as a normal array... should they
  // xxx be capitalized like the other parts of the API?
  push(item: T): void {
    this.inner.push(item);
    arrayAttribute.iterateAttrs(item, (attr: attrPrivate<unknown>) => {
      attr.addOutgoing(this);
      this.addIncoming(attr);
    });
    this.cleanAndNotify();
    this.update(this.inner);
  }

  // xxx fixme(iansmith) this should support the full api of slice()
  splice(start: number, deleteCount: number): T[] {
    const value = this.inner[start];
    if (value === undefined) {
      throw S5Err.NewError(S5Err.Messages.BadIndex, 'splice()');
    }
    arrayAttribute.iterateAttrs(value, (attr) => {
      attr.DropAttributeConnection(this);
    });
    const result = this.inner.splice(start, deleteCount);
    this.cleanAndNotify();
    this.update(this.inner);
    return result;
  }

  pop(): T {
    const value = this.inner.pop();
    if (value === undefined) {
      throw S5Err.NewError(S5Err.Messages.BadIndex, 'pop()');
    }
    arrayAttribute.iterateAttrs(value, (attr) => {
      attr.removeOutgoing(this);
      this.removeIncoming(attr);
    });
    this.cleanAndNotify();
    this.update(this.inner);
    return value;
  }

  /**
   * Get flattens the elements of the array of models. If you want
   * the attributes (not their values) use Raw().
   */
  public Get(): T[] {
    this.dirty = false;
    return decodeValue(this.inner);
  }

  /**
   * Raw returns the raw, unprocessed elements of the array. This is
   * not a copy, so any changes to this returned array will likely
   * be _very bad_.
   */
  public Raw(): T[] {
    this.dirty = false;
    return this.inner;
  }

  Set(a: T[]): void {
    const old = this.inner;
    this.inEdges.forEach((e) => {
      e.removeOutgoing(this);
    });
    this.inEdges = [];
    // use a sequence of pushes to get the new values into the array
    a.forEach((t: T) => {
      this.push(t);
    });
    this.inner = a;
    this.cleanAndNotify();
    this.update(this.inner, old);
  }

  index(i: number): T {
    return this.inner[i];
  }

  length(): number {
    return this.inner.length;
  }

  // eslint-disable-next-line class-methods-use-this
  AttributeName(): string {
    return 'ArrayAttribute';
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  map(fn: (t: T) => any): any[] {
    return this.inner.map(fn);
  }

  SetInputsAndFunc(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any,@typescript-eslint/no-unused-vars
    inputs: Array<attribute<any>>,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any,@typescript-eslint/no-unused-vars
    fn: (...a2: any[]) => T[]
  ): void {
    throw S5Err.NewError(
      S5Err.Messages.NotAllowed,
      `${this.DebugName()}:SetInputsAndFunc not allowed on array attribute`
    );
  }
}
