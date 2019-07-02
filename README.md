# sardines
协程池复用创建的协程，防止协程数过大。

## 安装
```` sh 
github.com/apoake/sardines
````

## 使用

```` go
p, _ := NewFixSizePools(10)  //固定十个协程
defer p.Close()

p.Summit(func() {  //提交任务
  fmt.Println("summit Func: ", time.Now())
})

f1 := func() (interface{}, error) {
  //Something todo
  return 1, nil
}
result, _ := p.SummitTask(f1)   // 提交任务，关心返回值
data, err := result.Get()   //调用Get方法阻塞，直到[f1]方法执行完成，获取返回值

f2 := func() (interface{}, error) {
  time.Sleep(5 * time.Second)
  return nil, errors.New("for test")
})
result, _ := p.SummitTask(f2)
//调用Get方法阻塞，等待[f2]方法执行完成，最多等待2秒
//超过2秒不返回结果，能获取到超时异常
data, err = result.GetTimed(2 * time.Second)    

````


