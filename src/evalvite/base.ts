import attribute from './attribute';

export enum topoMarkValues {
  none,
  temporary,
  Permanent,
}

// AttrPrivate is an interface that should not be needed by evalvite
// users, it is only useful in the implementation.  All attributes have
// these methods to support implementing the constraint graph and maintaining
// dirty bits.
export default interface attrPrivate<T> extends attribute<T> {
  markDirty(): void;

  update(oldValue?: T, newValue?: T): void;

  addOutgoing(target: attrPrivate<unknown>): void;

  removeOutgoing(target: attrPrivate<unknown>): void;

  getOutgoing(): Array<attrPrivate<unknown>>;

  mark(): topoMarkValues;

  setMark(t: topoMarkValues): void;

  // just for making debugging output nice
  AttributeName(): string;

  DropAttributeConnection(...other: attrPrivate<unknown>[]): void;
}
