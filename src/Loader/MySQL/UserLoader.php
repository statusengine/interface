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

namespace Statusengine\Loader\Mysql;


use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\UserLoaderInterface;

class UserLoader implements UserLoaderInterface {

    /**
     * @var \Statusengine\Backend\Mysql\MySQL
     */
    private $Backend;

    /**
     * UserLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    /**
     * @param string $username
     * @return array|bool
     */
    public function getUserByUsername($username){
        $query = $this->Backend->prepare(
            'SELECT * FROM statusengine_users WHERE username=? LIMIT 1'
        );
        $query->bindValue(1, $username);
        $result = $this->Backend->fetchAll($query);
        if(!empty($result)){
            return [
                'username' => $result[0]['username'],
                'password' => $result[0]['password']
            ];
        }

        return false;
    }

    public function getAllUsers() {
        $query = $this->Backend->prepare(
            'SELECT * FROM statusengine_users ORDER BY username ASC'
        );
        return $this->Backend->fetchAll($query);
    }

    /**
     * @param string $username
     * @param string $password
     * @return bool
     */
    public function addUser($username, $password) {
        if ($this->checkIfUsernameExists($username)) {
            throw new \RuntimeException(sprintf('Username %s already exists', $username));
        }

        $query = 'INSERT INTO statusengine_users (username, password)VALUES(?,?)';
        $query = $this->Backend->prepare($query);
        $query->bindValue(1, $username);
        $query->bindValue(2, $password);
        return $this->Backend->executeQuery($query);
    }

    /**
     * @param string $username
     * @param string $password
     * @return bool
     */
    public function changePassword($username, $password) {
        if (!$this->checkIfUsernameExists($username)) {
            throw new \RuntimeException(sprintf('Username %s does not exists', $username));
        }

        $query = 'UPDATE statusengine_users SET password=? WHERE username=?';
        $query = $this->Backend->prepare($query);
        $query->bindValue(1, $password);
        $query->bindValue(2, $username);
        return $this->Backend->executeQuery($query);
    }

    /**
     * @param string $username
     * @return bool
     */
    public function deleteUser($username) {
        if (!$this->checkIfUsernameExists($username)) {
            throw new \RuntimeException(sprintf('Username %s does not exists', $username));
        }

        $query = 'DELETE FROM statusengine_users WHERE username=?';
        $query = $this->Backend->prepare($query);
        $query->bindValue(1, $username);
        return $this->Backend->executeQuery($query);
    }

    /**
     * @param string $username
     * @return bool
     */
    private function checkIfUsernameExists($username) {
        $query = 'SELECT * FROM statusengine_users WHERE username=?';
        $query = $this->Backend->prepare($query);
        $query->bindValue(1, $username);
        if (!empty($this->Backend->fetchAll($query))) {
            return true;
        }
        return false;
    }

}
