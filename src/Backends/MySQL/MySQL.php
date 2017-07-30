<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */


namespace Statusengine\Backend\Mysql;

use Statusengine\Config;

class MySQL {

    /**
     * @var \Statusengine\Config
     */
    private $Config;

    /**
     * @var \PDO
     */
    protected $Connection;

    /**
     * MySQL constructor.
     * @param Config $Config
     */
    public function __construct(Config $Config) {
        $this->Config = $Config;
        return $this->connect();
    }


    /**
     * @return string
     */
    public function getDsn() {
        $config = $this->Config->getMysqlConfig();
        return sprintf(
            'mysql:host=%s;port=%s;dbname=%s',
            $config['host'],
            $config['port'],
            $config['database']
        );
    }

    /**
     * @return \PDO
     */
    public function connect() {
        $config = $this->Config->getMysqlConfig();
        $this->Connection = new \PDO($this->getDsn(), $config['username'], $config['password']);
        $this->Connection->setAttribute(\PDO::ATTR_ERRMODE, \PDO::ERRMODE_EXCEPTION);

        //$this->Connection->setAttribute(\PDO::ATTR_EMULATE_PREPARES, false);
        //$this->Connection->setAttribute(\PDO::ATTR_STRINGIFY_FETCHES, false);

        //Enable UTF-8
        $query = $this->Connection->prepare('SET NAMES utf8');
        $query->execute();

        return $this->Connection;
    }

    /**
     * @return \PDO
     */
    public function reconnect() {
        $this->Connection = null;
        return $this->connect();
    }


    /**
     * @param \PDOStatement $query
     * @return bool
     */
    public function executeQuery(\PDOStatement $query) {
        $result = false;
        try {
            $result = $query->execute();

        } catch (\Exception $Exception) {
            //todo implement error handling
        }
        return $result;
    }

    /**
     * @return \PDO
     */
    public function getConnection() {
        return $this->Connection;
    }


    /**
     * @param string $statement
     * @return \PDOStatement
     */
    public function prepare($statement) {
        $query = $this->Connection->prepare($statement);
        return $query;
    }

    /**
     * @param \PDOStatement $query
     * @return array
     */
    public function fetchAll(\PDOStatement $query) {
        $query->execute();
        return $query->fetchAll(\PDO::FETCH_ASSOC);
    }
}
