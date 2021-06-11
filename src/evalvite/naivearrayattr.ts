import attrPrivateImpl from './attrprivate';
import attribute from './attribute';
import * as S5Err from '../error';

export default class naiveArrayAttribute<T> extends attrPrivateImpl<T[]> {
  inner: T[] = [] as T[];

  constructor(debugName?: string) {
    super(debugName || `[naive array attribute]`);
  }

  push(item: T): void {
    this.inner.push(item);
    this.cleanAndNotify();
    this.update(this.inner);
  }

  pop(): T {
    const value = this.inner.pop();
    if (value === undefined) {
      throw new Error('unable to pop for NaiveAttrArray, result is undefined!');
    }
    this.cleanAndNotify();
    this.update(this.inner);
    return value;
  }

  Get(): T[] {
    this.dirty = false;
    return this.inner;
  }

  Set(a: T[]): void {
    this.inner = a;
    this.cleanAndNotify();
    this.update(this.inner);
  }

  index(i: number): T {
    return this.inner[i];
  }

  setIndex(i: number, v: T): void {
    this.inner[i] = v;
  }

  length(): number {
    return this.inner.length;
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  map(fn: (t: T) => any): any[] {
    return this.inner.map(fn);
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  decode(): any {
    return this.inner;
  }

  // eslint-disable-next-line class-methods-use-this
  AttributeName(): string {
    return 'NaiveArrayAttribute';
  }

  SetInputsAndFunc(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any,@typescript-eslint/no-unused-vars
    inputs: Array<attribute<any>>,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any,@typescript-eslint/no-unused-vars
    fn: (...a2: any[]) => T[]
  ): void {
    throw S5Err.NewError(
      S5Err.Messages.NotAllowed,
      `${this.DebugName()}:SetInputsAndFunc not allowed on naive array attribute`
    );
  }
}
