import couchdb

couch = couchdb.Server('http://admin:admin@83.128.10.233:5984/')

db = couch['weatherdata']

file = {'name' : 'Enschede', 'temp' : '19', 'humidity': '20'}
file2 = {'name' : 'Enschede', 'temp' : '20'}
file3 = {'name' : 'Enschede', 'temp' : '21'}
file4 = {'name' : 'Gronau', 'temp' : '22'}
file5 = {'name' : 'Enschede', 'temp' : '23'}
file6 = {'name' : 'Gronau', 'temp' : '24'}
file7 = {'name' : 'Enschede', 'temp' : '25'}
file8 = {'name' : 'Gronau', 'temp' : '26'}
file9 = {'name' : 'Enschede', 'temp' : '27'}
file10 = {'name' : 'Gronau', 'temp' : '30'}
db.save(file)
'''db.save(file2)
db.save(file3)
db.save(file4)
db.save(file5)
db.save(file6)
db.save(file7)
db.save(file8)
db.save(file9)
db.save(file10)'''
mango = { "selector": {
      "name": "Gronau"
   },
   "fields": ["temp"]
}
result = db.find(mango)
for row in result:
    print(row)

