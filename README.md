#Under developing

Table Construct
```
CREATE TABLE IF NOT EXISTS `counter` (
  `key` varchar(64) NOT NULL,
  `value` int(10) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

```
import(
	"go-sql-driver\mysql"
)

```