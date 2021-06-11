import attribute from './base';
import * as S5Conf from '../conf';
import basicAttribute from './basicattr';

let { IdCounter } = S5Conf.Vars;

export default class minMaxAttribute extends basicAttribute<number> {
  min: number;

  max: number;

  enforce = (v: number): number => {
    if (v < this.min) {
      return this.min;
    }
    if (v > this.max) {
      return this.max;
    }
    return v;
  };

  constructor(
    min: number,
    max: number,
    source: attribute<number>,
    name?: string | undefined
  ) {
    super(name || `[minmax ${IdCounter}]`);
    IdCounter += 1;
    this.SetInputsAndFunc([source], this.enforce);
    this.min = min;
    this.max = max;
    this.dirty = true; // because you have to compute it to have a cached copy
  }

  // eslint-disable-next-line class-methods-use-this
  AttributeName(): string {
    return `MinMax`;
  }
}
