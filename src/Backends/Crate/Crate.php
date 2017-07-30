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

namespace Statusengine\Backend\Crate;


use Crate\PDO\PDOStatement;
use Statusengine\Config;
use Crate\PDO\PDO as PDO;

class Crate {

    /**
     * @var Config
     */
    private $Config;

    /**
     * @var PDO
     */
    protected $Connection;


    /**
     * Crate constructor.
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
        $config = $this->Config->getCrateConfig();
        return sprintf('crate:%s', implode(',', $config));
    }

    /**
     * @return \Crate\PDO\PDO
     */
    public function connect() {
        $this->Connection = new PDO($this->getDsn(), null, null, null);
        $this->Connection->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
        return $this->Connection;
    }

    /**
     * @return \Crate\PDO\PDO
     */
    public function reconnect() {
        $this->Connection = null;
        return $this->connect();
    }

    /**
     * @return \Crate\PDO\PDO
     */
    public function getConnection() {
        return $this->Connection;
    }

    /**
     * @param PDOStatement $query
     * @return bool
     */
    public function executeQuery(PDOStatement $query) {
        $result = false;
        try {
            $result = $query->execute();

        } catch (\Exception $Exception) {
            print_r($Exception->errorInfo);
            //todo implement error handling
            /*
             * PHP Fatal error:  Uncaught exception 'GuzzleHttp\Exception\ConnectException' with message 'No more servers available, exception from last server: cURL error 28: Operation timed out after 5001 milliseconds with 0 bytes received (see http://curl.haxx.se/libcurl/c/libcurl-errors.html)' in /opt/statusengine-redis-5dadaf382f3e66ff3cd66a63df9b9f01df659860/redis/vendor/crate/crate-pdo/src/Crate/PDO/Http/Client.php:225
             */
        }
        return $result;
    }

    /**
     * @param PDOStatement $query
     * @return array
     */
    public function fetchAll(PDOStatement $query) {
        return $query->fetchAll(PDO::FETCH_ASSOC);
    }


    /**
     * @param string $statement
     * @return bool|\Crate\PDO\PDOStatement|\PDOStatement
     */
    public function prepare($statement) {
        return $this->Connection->prepare($statement);
    }


}
