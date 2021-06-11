import * as EV from '../../../src/evalvite'
import {expect} from 'chai';

// function getNanoSecTime() {
//   var hrTime = hrtime();
//   return hrTime[0] * 1000000000 + hrTime[1];
// }
function getMillisTime() {
  return Date.now();
}

const millisInSec= 1000;

describe('computes correct, to up to date values', ()=> {
  it('should handle updates to an array correctly', () => {
    // this is a "model" in our terminology... you must have an array
    // of "models" for array to work properly... you can't do this:
    // EV.ArrayAttr<SimpleAttribute<number>>
    type myModel = {
      [key: string]: any,
      v: EV.Attribute<number>,
    }
    //note: has to be a model, can't be just simpleAttribute<number>
    const a = EV.ArrayAttr<myModel>();

    for (let i = 0; i < 10; i++) {
      const rec = {v: EV.Simple<number>(i, `elem#${i}`)};
      a.push(rec);
    }
    // this is the type that would get passed to the state value in a
    // component or to a compute function, which is why we use it below.
    // it's the "flat" version of myModel above
    type flatMyModel = {
      v: number;
    }

    // usually, you don't want your deps on the elements of the array but
    // the array as a whole because you want to recompute if the number of
    // elements changes OR the value of some element changes
    const sum = EV.Computed<number>((models: flatMyModel[]) => {
      const values = models.map((m: flatMyModel) => m.v)
      return values.reduce((a: number, b: number) => a + b, 0);
    }, [a], 'sum')

    // for proving we don't do things we don't need to
    let computeCounts: number = 0;

    // have to get a dependency to the *array* here, not its content... but
    // that's really for avoiding the appearance of side-effects. If we *only*
    // had a dependency on the sum, we'd be ok, but then we'd be computing
    // our function over something
    const avg = EV.Computed<number>((sum: number, models: myModel[]) => {
      computeCounts = computeCounts + 1;
      return sum / models.length;
    }, [sum, a], 'avg')

    expect(sum.Get()).to.be.equal(45);
    expect(avg.Get()).to.be.equal(4.5);
    // test that the computed value is cached properly
    expect(sum.Get()/10).to.be.equal(4.5);
    expect(computeCounts).to.be.equal(1);

    const rec = {v: EV.Simple<number>(100)};
    a.push(rec);

    expect(sum.Get()).to.be.equal(145 /* 45 + 100 just added to end*/);
    expect(avg.Get()).to.be.equal(145 / 11); // safety first
    expect(computeCounts).to.be.equal(2); // two calls of sum


    computeCounts = 0;
    avg.Get();
    expect(computeCounts).to.be.equal(0);

    a.index(0).v.Set(33); // change elem 0 from 0 value to 33
    expect(sum.Get()).to.be.equal(145 + 33);
    expect(avg.Get()).to.be.equal((145 + 33) / 11); // safety first
    expect(computeCounts).to.be.equal(1);
  })
  // macbook pro, 2.6 GHz 6-Core Intel Core i7, node v14.13.1
  // iter=1
  // ~3000 marks/sec
  // ~3000 evals/sec
  // iter=50  (probably gets some more of the JIT compiler involved)
  // ~10000 marks/sec
  // ~3000 evals/sec
  it('should handle a long chain of dependencies correctly', () => {
    const start = EV.Simple<number>(0);
    let chainSize = 1500;
    let prev = start;
    for (let i = 0; i < chainSize; i = i + 1) {
      const curr = EV.Computed<number>((n: number) => n + 1,
        [prev], `chain-element-#${i}`);
      prev = curr;
    }
    //prev now points to elem chainSize in the chain
    expect(prev.Get()).to.be.equal(chainSize);

    let sumMark = 0;
    let sumEval = 0;
    const iters = 100;
    for (let i = 0; i < iters; i = i + 1) {
      const startMarking = getMillisTime();
      start.Set((i + 1) * chainSize);
      const finishMarking = getMillisTime();
      sumMark = sumMark + ((finishMarking - startMarking) / millisInSec);

      const startEval = getMillisTime()
      expect(prev.Get()).to.be.equal((i + 2) * chainSize);
      const finishEval = getMillisTime();
      sumEval = sumEval + ((finishEval - startEval) / millisInSec);
    }

    console.log("marking sum ",chainSize*iters/sumMark," marks/sec");
    console.log("eval sum ", chainSize*iters/sumEval, " evals/sec");

  }).slow(2000).timeout(5000);
  it('should correctly update length in naive array case', () => {
    const arr = EV.NaiveArray<number>()
    const lengthAttr = EV.Computed<number>( (naiveContent: number[])=>{
      return naiveContent.length;
    },[arr]);

    expect(lengthAttr.Get()).to.be.equal(0);

    arr.push(7);
    arr.push(8);
    arr.push(42);

    expect(lengthAttr.Get()).to.be.equal(3);

    arr.setIndex(2,0); // no effect on derived attributes!
    expect(lengthAttr.Get()).to.be.equal(3);

    const v=arr.pop();
    expect(v).to.be.equal(0);  // set above changed 42 to 0
    expect(lengthAttr.Get()).to.be.equal(2);

    arr.pop();
    arr.pop();
    expect(lengthAttr.Get()).to.be.equal(0);

  });
});